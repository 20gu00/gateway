package httpProxyServerMiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
)

//http服务限流
func HttpFlowLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverInterface, ok := ctx.Get("service")
		if !ok {
			common.Logger.Infof("未能从上下文中获取服务详细信息")
			ctx.Abort()
			return
		}

		serviceDetail := serverInterface.(*model.ServiceDetail)
		//网关开启服务端限流,针对服务的限流
		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := common.FlowLimiterHandler.GetLimiter(common.ServiceFlowStatPrefix+serviceDetail.BaseInfo.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				common.Logger.Infof("创建sevice的限流器失败(服务端限流)")
				ctx.Abort()
				return
			}

			//漏桶限流,拿到token请求即可通过
			if !serviceLimiter.Allow() { //Allow判断能否拿到token
				common.Logger.Infof(fmt.Sprintf("服务被限流了 %v", serviceDetail.AccessControl.ServiceFlowLimit))
				ctx.Abort()
				return
			}
		}

		//客户端限流(熔断),针对客户端ip的限流
		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := common.FlowLimiterHandler.GetLimiter(common.ServiceFlowStatPrefix+serviceDetail.BaseInfo.ServiceName+"_"+ctx.ClientIP(), float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				common.Logger.Infof("创建针对客户端ip的限流器失败")
				ctx.Abort()
				return
			}

			if !clientLimiter.Allow() {
				common.Logger.Infof(fmt.Sprintf("客户端ip被限流%v flow limit %v", ctx.ClientIP(), serviceDetail.AccessControl.ClientIPFlowLimit))
				ctx.Abort()
				return
			}

		}

		ctx.Next()
	}
}
