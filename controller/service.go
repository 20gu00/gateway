package controller

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type ServiceController struct{}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/list", service.ServiceList)
	group.GET("/delete", service.ServiceDelete)
	group.GET("/detail", service.ServiceDetail)
	group.GET("/stat", service.ServiceStat)
	group.POST("/add_http", service.ServiceAddHttp)
	group.POST("/update_http", service.ServiceUpdateHttp)
	group.POST("/add_tcp", service.ServiceAddTcp)
	group.POST("/update_tcp", service.ServiceUpdateTcp)
	group.POST("/add_grpc", service.ServiceAddGrpc)
	group.POST("/update_grpc", service.ServiceUpdateGrpc)
}

// ServiceList godoc
// @Summary service列表
// @Description service列表
// @Tags service管理
// @ID /service/list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_num query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/list [get]
func (s *ServiceController) ServiceList(ctx *gin.Context) {
	in := &dto.ServiceListInput{}
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{}
	serviceInfoList, total, err := serviceInfo.PageList(ctx, tx, in)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	outList := []dto.ServiceListItemOutput{}
	for _, listItem := range serviceInfoList {
		serviceDetail, err := listItem.ServiceDetail(ctx, tx, &listItem)
		if err != nil {
			middleware.ResponseError(ctx, 2003, err)
			return
		}

		//1、http前缀接入 clusterIP:clusterPort+path 2、http域名接入 domain 3、tcp、grpc接入 clusterIP:servicePort
		serviceAddr := "unknowServiceAddr"
		//网关的ip,端口
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		//http类型,接入类型,是否开启https
		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeHTTP && serviceDetail.HTTPRule.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}

		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeHTTP && serviceDetail.HTTPRule.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}

		//域名
		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeHTTP && serviceDetail.HTTPRule.RuleType == common.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}

		//
		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}

		if serviceDetail.ServiceInfo.LoadType == common.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}

		//实际的工作负载
		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		//流量统计,统计这个服务的流量
		counter, err := common.FlowCounterHandler.GetCounter(common.FlowServicePrefix + listItem.ServiceName) //服务流量统计前缀+服务名称
		if err != nil {
			middleware.ResponseError(ctx, 2004, err)
			return
		}

		outItem := dto.ServiceListItemOutput{
			ID:          listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         counter.QPS,
			Qpd:         counter.TotalCount,
			TotalNode:   len(ipList),
		}
		outList = append(outList, outItem)
	}
	out := &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	}
	middleware.ResponseSuccess(ctx, out)
}

// ServiceDelete godoc
// @Summary service删除
// @Description service删除
// @Tags service管理
// @ID /service/delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/delete [get]
func (s *ServiceController) ServiceDelete(ctx *gin.Context) {
	in := &dto.ServiceDeleteInput{}
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ID: in.ID}
	serviceInfo, err = serviceInfo.Find(ctx, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	serviceInfo.IsDelete = 1 //软删除
	if err := serviceInfo.Save(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "service删除成功")
}

// ServiceDetail godoc
// @Summary service详情
// @Description service详情
// @Tags service管理
// @ID /service/detail
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/detail [get]
func (s *ServiceController) ServiceDetail(ctx *gin.Context) {
	in := &dto.ServiceDeleteInput{}
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	//读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: in.ID}
	serviceInfo, err = serviceInfo.Find(ctx, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	//通过serviceinfo信息去夺标查询,拿到servicedetail
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	middleware.ResponseSuccess(ctx, serviceDetail)
}

// ServiceStat godoc
// @Summary service统计
// @Description service统计
// @Tags service管理
// @ID /service/stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/stat [get]
func (s *ServiceController) ServiceStat(ctx *gin.Context) {
	in := &dto.ServiceDeleteInput{}
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ID: in.ID}
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	//服务流量统计器,服务流量统计前缀+服务名称
	counter, err := common.FlowCounterHandler.GetCounter(common.FlowServicePrefix + serviceDetail.ServiceInfo.ServiceName)
	if err != nil {
		middleware.ResponseError(ctx, 2004, err)
		return
	}

	todayList := []int64{}
	currentTime := time.Now()
	//当日的数据
	for h := 0; h <= currentTime.Hour(); h++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), h, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime) //这个小时的数据
		todayList = append(todayList, hourData)      //追加到当天的数据列表中,按小时为粒度进行统计
	}

	yesterdayList := []int64{}
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	//昨日的数据
	for h := 0; h <= 23; h++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), h, 0, 0, 0, lib.TimeLocation) //时区在配置加载是初始化为"Asia/Shanghai"
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}

	middleware.ResponseSuccess(ctx, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}

// ServiceAddHttp godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags service管理
// @ID /service/add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/add_http [post]
func (s *ServiceController) ServiceAddHttp(ctx *gin.Context) {
	in := &dto.ServiceAddHttpInput{}
	//输入参数校验
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	//ip权重数目校验
	//接入网关的服务的地址,实际的工作负载,每条ip格式ip:port
	if len(strings.Split(in.IpList, ",")) != len(strings.Split(in.WeightList, ",")) {
		middleware.ResponseError(ctx, 2004, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	//这类型的添加,设计到多张表,往往组要servicedetail,修改多张表需要开启事务,这里使用本地事务
	tx = tx.Begin()

	//服务名称冲突校验
	serviceInfo := &dao.ServiceInfo{ServiceName: in.ServiceName}
	if _, err = serviceInfo.Find(ctx, tx, serviceInfo); err == nil {
		tx.Rollback()                                                       //出错就直接回滚
		middleware.ResponseError(ctx, 2002, errors.New("服务已存在,该服务名称已经被使用")) //判断新增的服务名称是否和已有的服务名称冲突
		return
	}

	httpUrl := &dao.HttpRule{RuleType: in.RuleType, Rule: in.Rule}
	//判断接入类型(前缀或者域名)和接入的前缀(路径)或域名是否冲突
	if _, err := httpUrl.Find(ctx, tx, httpUrl); err == nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("服务的接入前缀或域名已存在")) //网关统一代理全部后端实际的服务
		return
	}

	//存入service_info表
	serviceInfoDao := &dao.ServiceInfo{
		ServiceName: in.ServiceName,
		ServiceDesc: in.ServiceDesc,
	}
	if err := serviceInfoDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	//存入service_http_rule表
	httpDao := &dao.HttpRule{
		ServiceID:      serviceInfoDao.ID,
		RuleType:       in.RuleType,
		Rule:           in.Rule,
		NeedHttps:      in.NeedHttps,
		NeedStripUri:   in.NeedStripUri,
		NeedWebsocket:  in.NeedWebsocket,
		UrlRewrite:     in.UrlRewrite,
		HeaderTransfor: in.HeaderTransfor,
	}
	if err := httpDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	//存入service_access_control表
	accessControlDao := &dao.AccessControl{
		ServiceID:         serviceInfoDao.ID,
		OpenAuth:          in.OpenAuth,
		BlackList:         in.BlackList,
		WhiteList:         in.WhiteList,
		ClientIPFlowLimit: in.ClientipFlowLimit,
		ServiceFlowLimit:  in.ServiceFlowLimit,
	}
	if err := accessControlDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	//存入service_load_balance表
	loadbalanceDao := &dao.LoadBalance{
		ServiceID:              serviceInfoDao.ID,
		RoundType:              in.RoundType,
		IpList:                 in.IpList,
		WeightList:             in.WeightList,
		UpstreamConnectTimeout: in.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  in.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    in.UpstreamIdleTimeout,
		UpstreamMaxIdle:        in.UpstreamMaxIdle,
	}
	if err := loadbalanceDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	//全部成功那么就提交事务
	tx.Commit()
	middleware.ResponseSuccess(ctx, "添加http操作成功")
}

// ServiceUpdateHttp godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags service管理
// @ID /service/update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_http [post]
func (s *ServiceController) ServiceUpdateHttp(ctx *gin.Context) {
	//跟添加差不多,不过有些字段不能修改,比如服务名称
	in := &dto.ServiceUpdateHTTPInput{}
	if err := in.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	if len(strings.Split(in.IpList, ",")) != len(strings.Split(in.WeightList, ",")) {
		middleware.ResponseError(ctx, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	tx = tx.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: in.ServiceName}
	serviceInfo, err = serviceInfo.Find(ctx, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("service不存在"))
		return
	}

	//获取该service的servicedetail
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, errors.New("service不存在"))
		return
	}

	//service_info
	serviceInfoDao := serviceDetail.ServiceInfo
	serviceInfoDao.ServiceDesc = in.ServiceDesc
	if err := serviceInfoDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	//service_http_rule
	httpRuleDao := serviceDetail.HTTPRule
	httpRuleDao.NeedHttps = in.NeedHttps
	httpRuleDao.NeedStripUri = in.NeedStripUri
	httpRuleDao.NeedWebsocket = in.NeedWebsocket
	httpRuleDao.UrlRewrite = in.UrlRewrite
	httpRuleDao.HeaderTransfor = in.HeaderTransfor
	if err := httpRuleDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	//service_access_control
	accessControlDao := serviceDetail.AccessControl
	accessControlDao.OpenAuth = in.OpenAuth
	accessControlDao.BlackList = in.BlackList
	accessControlDao.WhiteList = in.WhiteList
	accessControlDao.ClientIPFlowLimit = in.ClientipFlowLimit
	accessControlDao.ServiceFlowLimit = in.ServiceFlowLimit
	if err := accessControlDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	//service_load_balance
	loadbalanceDao := serviceDetail.LoadBalance
	loadbalanceDao.RoundType = in.RoundType
	loadbalanceDao.IpList = in.IpList
	loadbalanceDao.WeightList = in.WeightList
	loadbalanceDao.UpstreamConnectTimeout = in.UpstreamConnectTimeout
	loadbalanceDao.UpstreamHeaderTimeout = in.UpstreamHeaderTimeout
	loadbalanceDao.UpstreamIdleTimeout = in.UpstreamIdleTimeout
	loadbalanceDao.UpstreamMaxIdle = in.UpstreamMaxIdle
	if err := loadbalanceDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(ctx, "更新http")
}

// ServiceAddTcp godoc
// @Summary 添加tcp服务
// @Description 添加tcp服务
// @Tags service管理
// @ID /service/add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/add_tcp [post]
func (admin *ServiceController) ServiceAddTcp(ctx *gin.Context) {
	in := &dto.ServiceAddTcpInput{}
	if err := in.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	infoSearch := &dao.ServiceInfo{
		ServiceName: in.ServiceName,
		IsDelete:    0,
	}
	//tx, err := lib.GetGormPool("default")
	if _, err := infoSearch.Find(ctx, lib.GORMDefaultPool, infoSearch); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("(tcp)服务名已经存在"))
		return
	}

	//判断端口是不是已经被使用了
	tcpRuleSearch := &dao.TcpRule{
		Port: in.Port,
	}
	if _, err := tcpRuleSearch.Find(ctx, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		middleware.ResponseError(ctx, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: in.Port,
	}
	if _, err := grpcRuleSearch.Find(ctx, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		middleware.ResponseError(ctx, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	if len(strings.Split(in.IpList, ",")) != len(strings.Split(in.WeightList, ",")) {
		middleware.ResponseError(ctx, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	serviceInfo := &dao.ServiceInfo{
		LoadType:    common.LoadTypeTCP,
		ServiceName: in.ServiceName,
		ServiceDesc: in.ServiceDesc,
	}
	if err := serviceInfo.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	loadBalanceDao := &dao.LoadBalance{
		ServiceID:  serviceInfo.ID,
		RoundType:  in.RoundType,
		IpList:     in.IpList,
		WeightList: in.WeightList,
		ForbidList: in.ForbidList,
	}
	if err := loadBalanceDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	httpRuleDao := &dao.TcpRule{
		ServiceID: serviceInfo.ID,
		Port:      in.Port,
	}
	if err := httpRuleDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	accessControlDao := &dao.AccessControl{
		ServiceID:         serviceInfo.ID,
		OpenAuth:          in.OpenAuth,
		BlackList:         in.BlackList,
		WhiteList:         in.WhiteList,
		WhiteHostName:     in.WhiteHostName,
		ClientIPFlowLimit: in.ClientIPFlowLimit,
		ServiceFlowLimit:  in.ServiceFlowLimit,
	}
	if err := accessControlDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2009, err)
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(ctx, "添加tcp成功")
	return
}

// ServiceUpdateTcp godoc
// @Summary 更新tcp服务
// @Description 更新tcp服务
// @Tags service管理
// @ID /service/update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_tcp [post]
func (s *ServiceController) ServiceUpdateTcp(ctx *gin.Context) {
	//tcp和grpc:添加端口,grpc:metadata
	in := &dto.ServiceUpdateTcpInput{}
	if err := in.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	if len(strings.Split(in.IpList, ",")) != len(strings.Split(in.WeightList, ",")) {
		middleware.ResponseError(ctx, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	serviceInfo := &dao.ServiceInfo{
		ID: in.ID,
	}
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, lib.GORMDefaultPool, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	seriverInfoDao := serviceDetail.ServiceInfo
	seriverInfoDao.ServiceDesc = in.ServiceDesc
	if err := seriverInfoDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	loadBalanceDao := &dao.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadBalanceDao = serviceDetail.LoadBalance
	}

	loadBalanceDao.ServiceID = serviceInfo.ID
	loadBalanceDao.RoundType = in.RoundType
	loadBalanceDao.IpList = in.IpList
	loadBalanceDao.WeightList = in.WeightList
	loadBalanceDao.ForbidList = in.ForbidList
	if err := loadBalanceDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, err)
		return
	}

	tcpRuleDao := &dao.TcpRule{}
	if serviceDetail.TCPRule != nil {
		tcpRuleDao = serviceDetail.TCPRule
	}

	tcpRuleDao.ServiceID = serviceInfo.ID
	tcpRuleDao.Port = in.Port
	if err := tcpRuleDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	accessControlDao := &dao.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControlDao = serviceDetail.AccessControl
	}

	accessControlDao.ServiceID = serviceInfo.ID
	accessControlDao.OpenAuth = in.OpenAuth
	accessControlDao.BlackList = in.BlackList
	accessControlDao.WhiteList = in.WhiteList
	accessControlDao.WhiteHostName = in.WhiteHostName
	accessControlDao.ClientIPFlowLimit = in.ClientIPFlowLimit
	accessControlDao.ServiceFlowLimit = in.ServiceFlowLimit
	if err := accessControlDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(ctx, "更新tcp服务成功")
	return
}

// ServiceAddGRPC godoc
// @Summary 添加grpc服务
// @Description 添加grpc服务
// @Tags service管理
// @ID /service/add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/add_grpc [post]
func (s *ServiceController) ServiceAddGrpc(ctx *gin.Context) {
	in := &dto.ServiceAddGrpcInput{}
	if err := in.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	search := &dao.ServiceInfo{
		ServiceName: in.ServiceName,
		IsDelete:    0,
	}
	if _, err := search.Find(ctx, lib.GORMDefaultPool, search); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("服务已经存在"))
		return
	}

	tcpSearch := &dao.TcpRule{
		Port: in.Port,
	}

	if _, err := tcpSearch.Find(ctx, lib.GORMDefaultPool, tcpSearch); err == nil {
		middleware.ResponseError(ctx, 2003, errors.New("服务端口被占用"))
		return
	}
	grpcSearch := &dao.GrpcRule{
		Port: in.Port,
	}
	if _, err := grpcSearch.Find(ctx, lib.GORMDefaultPool, grpcSearch); err == nil {
		middleware.ResponseError(ctx, 2004, errors.New("服务端口被占用"))
		return
	}

	if len(strings.Split(in.IpList, ",")) != len(strings.Split(in.WeightList, ",")) {
		middleware.ResponseError(ctx, 2005, errors.New("ip和权重数目不一致"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	infoDao := &dao.ServiceInfo{
		LoadType:    common.LoadTypeGRPC,
		ServiceName: in.ServiceName,
		ServiceDesc: in.ServiceDesc,
	}
	if err := infoDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	loadBalanceDao := &dao.LoadBalance{
		ServiceID:  infoDao.ID,
		RoundType:  in.RoundType,
		IpList:     in.IpList,
		WeightList: in.WeightList,
		ForbidList: in.ForbidList,
	}
	if err := loadBalanceDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	grpcDao := &dao.GrpcRule{
		ServiceID:      infoDao.ID,
		Port:           in.Port,
		HeaderTransfor: in.HeaderTransfor,
	}
	if err := grpcDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	accessControlDao := &dao.AccessControl{
		ServiceID:         infoDao.ID,
		OpenAuth:          in.OpenAuth,
		BlackList:         in.BlackList,
		WhiteList:         in.WhiteList,
		WhiteHostName:     in.WhiteHostName,
		ClientIPFlowLimit: in.ClientIPFlowLimit,
		ServiceFlowLimit:  in.ServiceFlowLimit,
	}
	if err := accessControlDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2009, err)
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(ctx, "添加grpc成功")
	return
}

// ServiceUpdateGrpc godoc
// @Summary 更新grpc服务
// @Description 更新grpc服务
// @Tags service管理
// @ID /service/update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_grpc [post]
func (s *ServiceController) ServiceUpdateGrpc(ctx *gin.Context) {
	in := &dto.ServiceUpdateGrpcInput{}
	if err := in.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(in.IpList, ",")) != len(strings.Split(in.WeightList, ",")) {
		middleware.ResponseError(ctx, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		ID: in.ID,
	}
	detail, err := service.ServiceDetail(ctx, lib.GORMDefaultPool, service)
	if err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	info := detail.ServiceInfo
	info.ServiceDesc = in.ServiceDesc
	if err := info.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, err)
		return
	}

	loadBalanceDao := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalanceDao = detail.LoadBalance
	}

	loadBalanceDao.ServiceID = info.ID
	loadBalanceDao.RoundType = in.RoundType
	loadBalanceDao.IpList = in.IpList
	loadBalanceDao.WeightList = in.WeightList
	loadBalanceDao.ForbidList = in.ForbidList
	if err := loadBalanceDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	grpcDao := &dao.GrpcRule{}
	if detail.GRPCRule != nil {
		grpcDao = detail.GRPCRule
	}
	grpcDao.ServiceID = info.ID
	//grpcRule.Port = params.Port
	grpcDao.HeaderTransfor = in.HeaderTransfor
	if err := grpcDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	accessControlDao := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControlDao = detail.AccessControl
	}

	accessControlDao.ServiceID = info.ID
	accessControlDao.OpenAuth = in.OpenAuth
	accessControlDao.BlackList = in.BlackList
	accessControlDao.WhiteList = in.WhiteList
	accessControlDao.WhiteHostName = in.WhiteHostName
	accessControlDao.ClientIPFlowLimit = in.ClientIPFlowLimit
	accessControlDao.ServiceFlowLimit = in.ServiceFlowLimit
	if err := accessControlDao.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(ctx, "更新tcp成功")
	return
}
