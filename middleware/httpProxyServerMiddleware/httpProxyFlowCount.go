package httpProxyServerMiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
)

//http服务的流量统计
func HttpFlowCountMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverInterface, ok := ctx.Get("service")
		if !ok {
			common.Logger.Infof("未能从上下文中获取服务详细信息 serviceDetail")
			ctx.Abort()
			return
		}

		serviceDetail := serverInterface.(*model.ServiceDetail)

		//大盘  服务service (统计近两天的流量数据,按小时为粒度)  (租户流量统计通过token来统计)
		totalCounter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
		if err != nil {
			common.Logger.Infof("获取统计器失败(用于统计全站的流量,用于大盘显示)")
			ctx.Abort()
			return
		}

		totalCounter.Increase() //累加

		//service的counter
		serviceCounter, err := common.FlowCounterHandler.GetCounter(common.ServiceFlowStatPrefix + serviceDetail.BaseInfo.ServiceName)
		if err != nil {
			common.Logger.Infof("获取统计器失败(用于统计service)")
			ctx.Abort()
			return
		}

		serviceCounter.Increase()
		ctx.Next() //传递给下一个中间件
	}
}
