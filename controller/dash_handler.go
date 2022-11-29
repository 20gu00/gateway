package controller

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func PanelDataHandler(c *gin.Context) {
	db := dao.DB
	serviceInfo := &model.ServiceInfo{}
	_, serviceNum, err := serviceInfo.PageList(db, &model.ServiceListInput{PageSize: 1, PageNum: 99999})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2000,
			"msg":  "获取服务的数目失败",
			"data": err.Error(),
		})
		return
	}
	//app := &model.Tenant{}
	tenant := &model.Tenant{}
	_, appNum, err := tenant.TenantList(db, &model.TenantListInput{PageNo: 1, PageSize: 99999})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2001,
			"msg":  "获取租户的数目失败",
			"data": err.Error(),
		})
		return
	}
	counter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 2002,
			"msg":  "创建统计器失败(用于统计整个站点的流量,大盘显示)",
			"data": err.Error(),
		})
		return
	}
	out := &model.PanelOutput{
		ServiceNum:      serviceNum,
		AppNum:          appNum,
		TodayRequestNum: counter.TotalCount,
		CurrentQPS:      counter.QPS,
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"code": 2003,
		"msg":  "大盘显示服务数目和租户数目和整个网关的qps(当前请求量)和qpd(当日请求量)",
		"data": err.Error(),
	})
}
