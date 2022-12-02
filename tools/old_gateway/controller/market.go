package controller

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

type MarketController struct{}

func MarketRegister(group *gin.RouterGroup) {
	service := &MarketController{}
	group.GET("/panel", service.PanelData)
	group.GET("/flow_stat", service.FlowStat)
	group.GET("/service_stat", service.ServiceStat)
}

// PanelData godoc
// @Summary 面板数据统计
// @Description 面板数据统计
// @Tags market
// @ID /market/panel
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelOutput} "success"
// @Router /market/panel [get]
func (service *MarketController) PanelData(c *gin.Context) {
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	_, serviceNum, err := serviceInfo.PageList(c, tx, &dto.ServiceListInput{PageSize: 1, PageNum: 1})
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	app := &dao.Tenant{}
	_, appNum, err := app.TenantList(c, tx, &dto.TenantListInput{PageNo: 1, PageSize: 1})
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	counter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	out := &dto.PanelOutput{
		ServiceNum:      serviceNum,
		AppNum:          appNum,
		TodayRequestNum: counter.TotalCount,
		CurrentQPS:      counter.QPS,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceStat godoc
// @Summary service按类型统计
// @Description service统计
// @Tags market
// @ID /market/service_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.MarketServiceStatOutput} "success"
// @Router /market/service_stat [get]
func (service *MarketController) ServiceStat(c *gin.Context) {
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	list, err := serviceInfo.GroupByLoadType(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	legend := []string{}
	for index, item := range list {
		name, ok := common.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(c, 2003, errors.New("load_type not found"))
			return
		}
		list[index].Name = name
		legend = append(legend, name)
	}
	out := &dto.MarketServiceStatOutput{
		Legend: legend,
		Data:   list,
	}
	middleware.ResponseSuccess(c, out)
}

// FlowStat godoc
// @Summary 流量统计
// @Description 流量统计
// @Tags market
// @ID /market/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /market/flow_stat [get]
func (service *MarketController) FlowStat(c *gin.Context) {
	counter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	todayList := []int64{}
	currentTime := time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayList = append(todayList, hourData)
	}

	yesterdayList := []int64{}
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}
	middleware.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}
