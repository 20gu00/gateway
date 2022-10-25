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

//更新逻辑,可以是使用service_id,但对于用户来说是使用service_name,servicename不应该被修改
//这些可以有前端处理,但我偏好后端实现服务名而不是id处理更新逻辑

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

	//service := &dao.ServiceInfo{
	//	ID: in.ID,
	//}

	serviceInfo := &dao.ServiceInfo{ServiceName: in.ServiceName}
	serviceInfo, err := serviceInfo.Find(ctx, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2009, errors.New("service不存在(通过servicename查找service)"))
		return
	}

	detail, err := serviceInfo.ServiceDetail(ctx, lib.GORMDefaultPool, serviceInfo)
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
