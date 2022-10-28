package controller

import (
	"errors"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"strings"
)

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

	//serviceInfo := &dao.ServiceInfo{
	//	ID: in.ID,
	//}

	serviceInfo := &dao.ServiceInfo{ServiceName: in.ServiceName}
	serviceInfo, err := serviceInfo.Find(ctx, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2009, errors.New("service不存在(通过servicename查找service)"))
		return
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
