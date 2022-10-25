package http_proxy_middleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//统计项 1 全站 2 服务 3 租户
		totalCounter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		totalCounter.Increase() //累加

		//dayCount, _ := totalCounter.GetDayData(time.Now())
		//fmt.Printf("totalCounter qps:%v,dayCount:%v", totalCounter.QPS, dayCount)
		serviceCounter, err := common.FlowCounterHandler.GetCounter(common.FlowServicePrefix + serviceDetail.ServiceInfo.ServiceName)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		serviceCounter.Increase()

		//dayServiceCount, _ := serviceCounter.GetDayData(time.Now())
		//fmt.Printf("serviceCounter qps:%v,dayCount:%v", serviceCounter.QPS, dayServiceCount)
		c.Next()
	}
}
