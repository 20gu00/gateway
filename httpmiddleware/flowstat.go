package httpmiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//http服务的流量统计
func HttpFlowCountMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("未能从上下文中获取服务详细信息"))
			ctx.Abort()
			return
		}

		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//站点(大盘)  服务  租户(都是统计近两天的流量数据,按小时为粒度)(租户流量统计通过token来统计)
		totalCounter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
		if err != nil {
			middleware.ResponseError(ctx, 4001, err)
			ctx.Abort()
			return
		}

		totalCounter.Increase() //累加

		//dayCount, _ := totalCounter.GetDayData(time.Now())
		serviceCounter, err := common.FlowCounterHandler.GetCounter(common.FlowServicePrefix + serviceDetail.ServiceInfo.ServiceName)
		if err != nil {
			middleware.ResponseError(ctx, 4001, err)
			ctx.Abort()
			return
		}

		serviceCounter.Increase()
		ctx.Next() //传递给下一个中间件
	}
}
