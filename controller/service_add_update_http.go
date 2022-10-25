package controller

import (
	"errors"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"strings"
)

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
	//
	serviceInfo := &dao.ServiceInfo{ServiceName: in.ServiceName}
	serviceInfo, err = serviceInfo.Find(ctx, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("service不存在(通过servicename查找service)"))
		return
	}

	//获取该service的servicedetail
	serviceDetail, err := serviceInfo.ServiceDetail(ctx, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, errors.New("service不存在(获取servicedetail失败)"))
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
