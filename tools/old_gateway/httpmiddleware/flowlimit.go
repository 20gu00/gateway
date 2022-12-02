package httpmiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//http服务限流
func HttpFlowLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("未能从上下文中获取服务详细信息"))
			ctx.Abort()
			return
		}

		serviceDetail := serverInterface.(*dao.ServiceDetail)
		//网关开启服务端限流,针对服务的限流
		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := common.FlowLimiterHandler.GetLimiter(common.FlowServicePrefix+serviceDetail.ServiceInfo.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				middleware.ResponseError(ctx, 5001, err)
				ctx.Abort()
				return
			}

			//漏桶限流,拿到token请求即可通过
			if !serviceLimiter.Allow() { //Allow判断能否拿到token
				middleware.ResponseError(ctx, 5002, errors.New(fmt.Sprintf("服务被限流了 %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				ctx.Abort()
				return
			}
		}

		//客户端限流(熔断),针对客户端ip的限流
		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := common.FlowLimiterHandler.GetLimiter(common.FlowServicePrefix+serviceDetail.ServiceInfo.ServiceName+"_"+ctx.ClientIP(), float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				middleware.ResponseError(ctx, 5003, err)
				ctx.Abort()
				return
			}

			if !clientLimiter.Allow() {
				middleware.ResponseError(ctx, 5002, errors.New(fmt.Sprintf("客户端ip被限流%v flow limit %v", ctx.ClientIP(), serviceDetail.AccessControl.ClientIPFlowLimit)))
				ctx.Abort()
				return
			}

		}
		ctx.Next()
	}
}
