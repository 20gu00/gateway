package controller

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/dto"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"time"
)

// ServiceStat godoc
// @Summary service统计(流量)
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
