package model

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type ServiceInfo struct {
	gorm.Model
	LoadType    int    `json:"load_type"`
	ServiceName string `json:"service_name"`
	ServiceDesc string `json:"service_desc"`
	IsDelete    int    `json:"is_delete"`
}

func (*ServiceInfo) TableName() string {
	return "service_info"
}

//列出service信息(base表)
func (s *ServiceInfo) PageList(db *gorm.DB, input *ServiceListInput) ([]ServiceInfo, int64, error) {
	total := int64(0)
	list := []ServiceInfo{}
	offset := (input.PageNum - 1) * input.PageSize

	query := db.Table(s.TableName()).Where("is_delete=0")
	if input.Info != "" {
		query = query.Where("service_name like ? or service_desc like ?", "%"+input.Info+"%", "%"+input.Info+"%")
	}

	//原生sql语句limit 10  limit 1,10  asc升 desc降
	if err := query.Limit(input.PageSize).Offset(offset).Order("id desc").Find(&list); err != nil {
		return nil, 0, err.Error
	}

	query.Limit(input.PageSize).Offset(offset).Count(&total)
	return list, total, nil
}

func (s *ServiceInfo) Find(c *gin.Context, db *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	result := &ServiceInfo{}
	//根据该结构体跟数据库表的映射关系(search)不用Table(s.TableName())也行
	row := db.Table(s.TableName()).Where(search).First(result).RowsAffected
	if row != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2000,
			"msg":  "该service不存在",
			"data": row,
		})
		return nil, errors.New("该service不存在")
	}

	return result, nil
}

type ServiceDetail struct {
	BaseInfo *ServiceInfo
	Http     *Service_http
	//Tcp           *Service_tcp
	AccessControl *AccessControl
	LoadBalance   *LoadBalance
}

//整合service相关的所有的表,serviceInfo是基本表,其他表与serviceInfo的id进行联结的
func (s *ServiceInfo) ServiceDetail(db *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	//service_name通过参数校验确定唯一,提供逻辑给前端使用,通过service_name来获取该service的信息
	if search.ServiceName == "" {
		info := new(ServiceInfo)
		//结构体指针
		if tx := db.Table(s.TableName()).Where(search).Find(info); tx.Error != nil {
			return nil, tx.Error
		}
		search = info
	}

	//Http表
	httpSearch := &Service_http{
		ServiceId: int(search.ID),
	}
	httpResult := new(Service_http)
	if tx := db.Table(httpSearch.TableName()).Where(httpSearch).Find(httpResult); tx.Error != nil {
		return nil, tx.Error
	}

	////Tcp表
	//tcpSearch := &Service_tcp{
	//	ServiceId: int(search.ID),
	//}
	//tcpResult := new(Service_tcp)
	//if tx := db.Table(tcpSearch.TableName()).Where(tcpSearch).Find(tcpResult); tx.Error != nil {
	//	return nil, tx.Error
	//}

	//access_control
	acSearch := &AccessControl{
		ServiceId: int(search.ID),
	}
	acResult := new(AccessControl)
	if tx := db.Table(acSearch.TableName()).Where(acSearch).Find(acResult); tx.Error != nil {
		return nil, tx.Error
	}

	//loadbalance
	lbSearch := &LoadBalance{
		ServiceId: int(search.ID),
	}
	lbResult := new(LoadBalance)
	if tx := db.Table(lbSearch.TableName()).Where(lbSearch).Find(lbResult); tx.Error != nil {
		return nil, tx.Error
	}

	serviceDetail := &ServiceDetail{
		BaseInfo: search,
		Http:     httpResult,
		//Tcp:           tcpResult,
		LoadBalance:   lbResult,
		AccessControl: acResult,
	}
	return serviceDetail, nil
}

type Service_http struct {
	ID             int    `gorm:"primary_key" json:"id"`
	ServiceId      int    `json:"service_id"`
	RuleType       int    `json:"rule_type"`
	Rule           string `json:"rule"`
	NeedHttps      int    `json:"need_https"`
	NeedStrip_uri  int    `json:"need_strip_uri"`
	NeedWebsocket  int    `json:"need_websocket"`
	UrlRewrite     string `json:"url_rewrite"`
	HeaderTransfor string `json:"header_transfor"`
}

func (*Service_http) TableName() string {
	return "service_http"
}

type LoadBalance struct {
	ID                     int    `gorm:"primary_key" json:"id"`
	ServiceId              int    `json:"service_id"`
	CheckMethod            int    `json:"check_method"`
	CheckTimeout           int    `json:"check_timeout"`
	CheckInterval          int    `json:"check_interval"`
	RoundType              int    `json:"round_type"`
	IpList                 string `json:"ip_list"`
	WeightList             string `json:"weight_list"`
	ForbidList             string `json:"forbid_list"`
	UpstreamConnectTimeOut int    `json:"upstream_connect_timeout"`
	UpstreamHeaderTimeOut  int    `json:"upstream_header_timeout"`
	UpstreamIdleTimeOut    int    `json:"upstream_idle_timeout"`
	UpstreamMaxIdle        int    `json:"upstream_max_idle"`
}

func (*LoadBalance) TableName() string {
	return "loadbalance"
}

func (l *LoadBalance) GetIpList() []string {
	return strings.Split(l.IpList, ",")
}

func (l *LoadBalance) GetWeightList() []string {
	return strings.Split(l.WeightList, ",")
}

type AccessControl struct {
	ID                int    `gorm:"primary_key" json:"id"`
	ServiceId         int    `json:"service_id"`
	OpenAuth          int    `json:"open_auth"`
	BlackList         string `json:"black_list"`
	WhiteList         string `json:"white_list"`
	WhiteHostName     string `json:"white_host_name"` //`json:"white_host_name" gorm:"column:white_host_name" description:"白名单host"`
	ClientIPFlowLimit int    `json:"clientip_flow_limit"`
	ServiceFlowLimit  int    `json:"service_flow_limit"`
}

func (*AccessControl) TableName() string {
	return "access_control"
}

//input output

//serviceList
type ServiceListInput struct {
	Info     string `json:"info"`      //搜索关键词  `json:"info" form:"info" comment:"关键词" example:"" validate:""`
	PageNum  int    `json:"page_num"`  //页数
	PageSize int    `json:"page_size"` //每页条数

}

type ServiceListItemOutput struct {
	ID          int64  `json:"id"`           //id
	ServiceName string `json:"service_name"` //服务名称
	ServiceDesc string `json:"service_desc"` //服务描述
	LoadType    int    `json:"load_type"`    //类型
	ServiceAddr string `json:"service_addr"` //服务地址
	Qps         int64  `json:"qps"`          //qps  每秒访问量每秒请求率
	Qpd         int64  `json:"qpd"`          //qpd  每天
	TotalNode   int    `json:"total_node"`   //节点数
}

type ServiceListOutput struct {
	ServiceList []ServiceListItemOutput
	Total       int
}

//serviceDelete
type ServiceDeleteInput struct {
	ID int64 `json:"id"` //服务ID
}

//serviceDetail
type ServiceDetailInput struct {
	ID int64 `json:"id"` //服务ID
}

//serviceStat
type ServiceStatInput struct {
	ID int `json:"id"`
}

//serviceAddHttp
type ServiceAddHttpInput struct {
	//service_info
	ServiceName string `json:"service_name"` //服务名
	ServiceDesc string `json:"service_desc"` //服务描述

	//service_http_rule
	RuleType       int    `json:"rule_type" `      //接入类型,0是前缀
	Rule           string `json:"rule"`            //域名或者前缀(路径/add)
	NeedHttps      int    `json:"need_https"`      //支持https
	NeedStripUri   int    `json:"need_strip_uri"`  //启用strip_uri,注意这个功能和url重写冲突性
	NeedWebsocket  int    `json:"need_websocket"`  //是否支持websocket
	UrlRewrite     string `json:"url_rewrite"`     //url重写功能
	HeaderTransfor string `json:"header_transfor"` //header转换

	//service_access_control
	OpenAuth          int    `json:"open_auth"`           //关键词
	BlackList         string `json:"black_list"`          //黑名单ip
	WhiteList         string `json:"white_list"`          //白名单ip
	ClientipFlowLimit int    `json:"clientip_flow_limit"` //是否开启客户端ip限流
	ServiceFlowLimit  int    `json:"service_flow_limit"`  //服务端限流

	//service_load_balance
	RoundType              int    `json:"round_type"`               //轮询方式
	IpList                 string `json:"ip_list"`                  //ip列表
	WeightList             string `json:"weight_list"`              //权重列表
	UpstreamConnectTimeout int    `json:"upstream_connect_timeout"` //建立连接超时, 单位s
	UpstreamHeaderTimeout  int    `json:"upstream_header_timeout"`  //链接最大空闲时间, 单位s
	UpstreamMaxIdle        int    `json:"upstream_max_idle"`        //最大空闲链接数
	UpstreamIdleTimeout    int    `json:"upstream_idle_timeout"`    //链接最大空闲时间, 单位s
}

type ServiceUpdateHttpInput struct {
	ID          int64  `json:"id"`            //服务ID
	ServiceName string `json:"service_name" ` //服务名
	ServiceDesc string `json:"service_desc"`  //服务描述

	RuleType       int    `json:"rule_type"`       //接入类型
	Rule           string `json:"rule"`            //域名或者前缀
	NeedHttps      int    `json:"need_https"`      //支持https
	NeedStripUri   int    `json:"need_strip_uri"`  //启用strip_uri
	NeedWebsocket  int    `json:"need_websocket"`  //是否支持websocket
	UrlRewrite     string `json:"url_rewrite"`     //url重写功能
	HeaderTransfor string `json:"header_transfor"` //header转换

	OpenAuth          int    `json:"open_auth"`           //关键词
	BlackList         string `json:"black_list"`          //黑名单ip
	WhiteList         string `json:"white_list"`          //白名单ip
	ClientipFlowLimit int    `json:"clientip_flow_limit"` //客户端ip限流
	ServiceFlowLimit  int    `json:"service_flow_limit"`  //服务端限流

	RoundType              int    `json:"round_type"`               //轮询方式
	IpList                 string `json:"ip_list"`                  //ip列表
	WeightList             string `json:"weight_list"`              //权重列表
	UpstreamConnectTimeout int    `json:"upstream_connect_timeout"` //建立连接超时, 单位s
	UpstreamHeaderTimeout  int    `json:"upstream_header_timeout"`  //获取header超时, 单位s
	UpstreamIdleTimeout    int    `json:"upstream_idle_timeout"`    //链接最大空闲时间, 单位s
	UpstreamMaxIdle        int    `json:"upstream_max_idle"`        //最大空闲链接数
}

type ServiceStatOutput struct {
	Today     []int64 `json:"today" form:"today" comment:"今日流量" example:"" validate:""`         //列表
	Yesterday []int64 `json:"yesterday" form:"yesterday" comment:"昨日流量" example:"" validate:""` //列表
}
