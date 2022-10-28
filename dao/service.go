package dao

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dto"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetail //原生map非线程安全,需要加锁,写入map如果有县城读取map出错(map查询效率高)
	ServiceSlice []*ServiceDetail          //减少锁的开销
	Locker       sync.RWMutex
	init         sync.Once
	err          error //Once中的错误
}

type ServiceDetail struct {
	ServiceInfo   *ServiceInfo   `json:"serviceInfo" description:"service的基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule表"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule表"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance表"`
	AccessControl *AccessControl `json:"access_control" description:"access_control表"`
}

var ServiceManagerHandler *ServiceManager

//调用这个包时直接初始化
func init() {
	ServiceManagerHandler = NewServiceManager()
}

//新建个servicemanager,用来处理service
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

//获取tcp服务
func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	list := []*ServiceDetail{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		//过滤出tcp的服务写入服务详情列表
		if tempItem.ServiceInfo.LoadType == common.LoadTypeTCP {
			list = append(list, tempItem)
		}

	}
	return list
}

//http服务的接入方式
func (s *ServiceManager) HTTPAccessMode(ctx *gin.Context) (*ServiceDetail, error) {
	//前缀 /abc 域名匹配(c.Request.URL.Path)  www.cjq.com(c.Request.Host)
	host := ctx.Request.Host
	host = host[0:strings.Index(host, ":")] //会连端口一块获取,需要切割下

	path := ctx.Request.URL.Path
	for _, serviceItem := range s.ServiceSlice {
		//负载均衡类型判断请求的类型
		if serviceItem.ServiceInfo.LoadType != common.LoadTypeHTTP {
			continue //这里处理的是http,不是http就继续下一个循环
		}

		//判断http的接入类型
		if serviceItem.HTTPRule.RuleType == common.HTTPRuleTypeDomain {
			if serviceItem.HTTPRule.Rule == host { //匹配成功,找到要请求的服务,返回该服务的servicedetail
				return serviceItem, nil
			}
		}

		if serviceItem.HTTPRule.RuleType == common.HTTPRuleTypePrefixURL {
			//测试path的前缀是不是以serviceItem.HTTPRule.Rule开头
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("请求未能匹配到任何网关所管理的服务")
}

//将数据从数据库中加载到内存,一次加载
func (s *ServiceManager) LoadOnce() error {
	s.init.Do(func() { //sync.Once调用
		serviceInfo := &ServiceInfo{}
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder()) //新的上下文,测试使用
		tx, err := lib.GetGormPool("default")                   //获取连接池               //从default这个pool中获取连接
		if err != nil {
			s.err = err
			return
		}

		params := &dto.ServiceListInput{PageNum: 1, PageSize: 99999}
		serviceInfoList, _, err := serviceInfo.PageList(ctx, tx, params)
		if err != nil {
			s.err = err
			return
		}

		s.Locker.Lock() //原生的map不是并发安全的,需要加锁,比如你在写入map,如果有额外的线程在读取map那么就会报错
		defer s.Locker.Unlock()
		for _, serviceInfoListItem := range serviceInfoList { //range值传递,右边的表达式只计算一次,变量只创建一次,也就是地址不变,如果后续有使用这个指针,那么到最后range完成,使用该指针的值最后都一样都是这个变量的值
			tmpItem := serviceInfoListItem //值传递
			serviceDetail, err := tmpItem.ServiceDetail(ctx, tx, &tmpItem)
			if err != nil {
				s.err = err
				return
			}

			//写入servicedetail的map和slice
			s.ServiceMap[serviceInfoListItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})
	return s.err //s中的err,once的错误
}
