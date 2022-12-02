package manager

import (
	"errors"
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"strings"
	"sync"
)

var ServiceManagerHandler *ServiceManager

//专门用于处理service的对象
type ServiceManager struct {
	ServiceMap   map[string]*model.ServiceDetail //原生map非线程安全,需要加锁,写入map如果有线程读取map出错(map查询效率高)
	ServiceSlice []*model.ServiceDetail          //减少锁的开销
	Locker       sync.RWMutex                    //给原生map加个锁
	init         sync.Once
	err          error //Once中的错误
}

//实例化ServiceManager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*model.ServiceDetail{},
		ServiceSlice: []*model.ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

//调用这个包时直接初始化(只加载一次,如果数据库数据变更,所以设置一定时间重载)
func init() {
	ServiceManagerHandler = NewServiceManager()
}

//获取tcp服务
func (s *ServiceManager) GetTcpServiceList() []*model.ServiceDetail {
	list := []*model.ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		//过滤出tcp的服务写入服务详情列表
		if tempItem.BaseInfo.LoadType == common.LoadTypeTCP {
			list = append(list, tempItem)
		}

	}
	return list
}

//http服务的接入方式并返回service的serviceDetail
func (s *ServiceManager) HttpAccessMode(ctx *gin.Context) (*model.ServiceDetail, error) {
	//前缀 /abc 域名匹配(c.Request.URL.Path)  www.cjq.com(c.Request.Host)
	//要访问的服务器地址(访问了网关)

	//案例:http.ListenAndServe("localhost:9999", nil)
	//使用 curl 命令访问：curl http://localhost:9999/a/b/c  (ip:port或者domain:port)(domain得解析到网关)(/a/ www.aaa.com)
	//r.Host 是 localhost:9999，(ip:port)
	//r.URL.Host 是空字符串，
	//r.URL.Path 是 /a/b/c

	host := ctx.Request.Host
	host = host[0:strings.Index(host, ":")] //会连端口一块获取,返回str(:)在host中第一次出现的位置(字符串的索引包括空格)(如果找不到则返回-1；如果str为空，则返回0)

	path := ctx.Request.URL.Path
	for _, serviceItem := range s.ServiceSlice {
		//负载均衡类型判断请求的类型
		if serviceItem.BaseInfo.LoadType != common.LoadTypeHTTP {
			continue //这里处理的是http,不是http就继续下一个循环
		}

		//判断http的接入类型
		if serviceItem.Http.RuleType == common.HTTPRuleTypeDomain {
			if serviceItem.Http.Rule == host { //匹配成功,找到要请求的服务,返回该服务的servicedetail
				return serviceItem, nil
			}
		}

		if serviceItem.Http.RuleType == common.HTTPRuleTypePrefixURL {
			//测试path的前缀是不是以serviceItem.HTTPRule.Rule开头
			if strings.HasPrefix(path, serviceItem.Http.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("请求未能匹配到任何网关可以处理的的服务类型(支持http请求的前缀和域名方式)")
}

//将数据从数据库中加载到内存,一次加载,service
func (s *ServiceManager) LoadOnce() error {
	s.init.Do(func() { //sync.Once调用(只运行一次)
		serviceInfo := &model.ServiceInfo{}
		//ctx, _ := gin.CreateTestContext(httptest.NewRecorder()) //新的上下文,测试使用
		db := dao.DB

		search := &model.ServiceListInput{PageNum: 1, PageSize: 99999} //全部列出
		serviceInfoList, _, err := serviceInfo.PageList(db, search)
		if err != nil {
			s.err = err
			fmt.Println(err.Error())
			return
		}

		s.Locker.Lock() //操作map前上锁
		defer s.Locker.Unlock()

		for _, serviceInfoListItem := range serviceInfoList { //range值传递,右边的表达式只计算一次,变量只创建一次,也就是地址不变,如果后续有使用这个指针,那么到最后range完成,使用该指针的值最后都一样都是这个变量的值
			tmpItem := serviceInfoListItem //值传递
			serviceDetail, err := tmpItem.ServiceDetail(db, &tmpItem)
			if err != nil {
				s.err = err
				fmt.Println(err.Error())
				return
			}

			//写入servicedetail的map和slice
			s.ServiceMap[serviceInfoListItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})
	return s.err //s中的err,once的错误
}
