package controller

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//APPControllerRegister admin路由注册
func APPRegister(router *gin.RouterGroup) {
	admin := TenantController{}
	router.GET("/list", admin.TenantList)
	router.GET("/detail", admin.TenantDetail)
	router.GET("/stat", admin.TenantStat)
	router.GET("/delete", admin.TenantDelete)
	router.POST("/add", admin.TenantAdd)
	router.POST("/update", admin.TenantUpdate)
}

type TenantController struct{}

// TenantList godoc
// @Summary 网关租户列表
// @Description 网关租户列表
// @Tags tenant管理
// @ID /tenant/list
// @Accept  json
// @Produce  json
// @Param info query string false "搜索关键词"
// @Param page_size query string true "每页数目"
// @Param page_num query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.TenantListOutput} "success"
// @Router /tenant/list [get]
func (t *TenantController) TenantList(ctx *gin.Context) {
	in := &dto.TenantListInput{}
	if err := in.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	info := &dao.Tenant{}
	list, total, err := info.TenantList(ctx, lib.GORMDefaultPool, in)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	outputList := []dto.TenantListItemOutput{}
	for _, item := range list {
		appCounter, err := common.FlowCounterHandler.GetCounter(common.FlowAppPrefix + item.AppID)
		if err != nil {
			middleware.ResponseError(ctx, 2003, err)
			ctx.Abort()
			return
		}

		outputList = append(outputList, dto.TenantListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := dto.TenantListOutput{
		List:  outputList,
		Total: total,
	}
	middleware.ResponseSuccess(ctx, output)
	return
}

// TenantDetail godoc
// @Summary 网关租户详情
// @Description 网关租户详情
// @Tags tenant管理
// @ID /tenant/detail
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.Tenant} "success"
// @Router /tenant/detail [get]
func (t *TenantController) TenantDetail(ctx *gin.Context) {
	params := &dto.TenantDetailInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	search := &dao.Tenant{
		ID: params.ID,
	}
	detail, err := search.Find(ctx, lib.GORMDefaultPool, search)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	middleware.ResponseSuccess(ctx, detail)
	return
}

// TenantDelete godoc
// @Summary 网关租户删除
// @Description 网关租户删除
// @Tags tenant管理
// @ID /tenant/delete
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /tenant/delete [get]
func (admin *TenantController) TenantDelete(c *gin.Context) {
	params := &dto.TenantDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	search := &dao.Tenant{
		ID: params.ID,
	}
	info, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	info.IsDelete = 1
	if err := info.Save(c, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	middleware.ResponseSuccess(c, "网关租户删除成功")
	return
}

// TenantAdd godoc
// @Summary 网关租户添加
// @Description 网关租户添加
// @Tags tenant管理
// @ID /tenant/add
// @Accept  json
// @Produce  json
// @Param body body dto.TenantAddHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /tenant/add [post]
func (admin *TenantController) TenantAdd(c *gin.Context) {
	params := &dto.TenantAddHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.Tenant{
		AppID: params.AppID,
	}
	if _, err := search.Find(c, lib.GORMDefaultPool, search); err == nil {
		middleware.ResponseError(c, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}

	if params.Secret == "" {
		//生成secret
		params.Secret = common.MD5(params.AppID)
	}

	tx := lib.GORMDefaultPool
	info := &dao.Tenant{
		AppID:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps:      params.Qps,
		Qpd:      params.Qpd,
	}
	if err := info.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	middleware.ResponseSuccess(c, "添加网关的租户成功")
	return
}

// TenantUpdate godoc
// @Summary 网关租户更新
// @Description 网关租户更新
// @Tags tenant管理
// @ID /tenant/update
// @Accept  json
// @Produce  json
// @Param body body dto.TenantUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /tenant/update [post]
func (admin *TenantController) TenantUpdate(c *gin.Context) {
	params := &dto.TenantUpdateInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	search := &dao.Tenant{
		ID: params.ID,
	}
	info, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	if params.Secret == "" {
		params.Secret = common.MD5(params.AppID)
	}
	info.Name = params.Name
	info.Secret = params.Secret
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(c, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "修改网关租户配置成功")
	return
}

// TenantStat godoc
// @Summary 网关租户统计
// @Description 网关租户统计
// @Tags tenant管理
// @ID /tenant/stat
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /tenant/stat [get]
