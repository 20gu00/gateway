package controller

import (
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
)

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
