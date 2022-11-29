package httpProxyServerMiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
)

//基于jwt的限流(租户 客户端 访问服务)
func HttpJwtFlowLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantInterface, ok := ctx.Get("app")
		if !ok {
			ctx.Next()
			return
		}

		tenantInfo := tenantInterface.(*model.Tenant)
		if tenantInfo.Qps > 0 {
			clientLimiter, err := common.FlowLimiterHandler.GetLimiter(common.FlowTenantPrefix+tenantInfo.AppId+"_"+ctx.ClientIP(), float64(tenantInfo.Qps))
			if err != nil {
				common.Logger.Infof("创建流量统计器失败(用于统计租户访问服务的流量)")
				ctx.Abort()
				return
			}

			if !clientLimiter.Allow() {
				common.Logger.Infof("租户限流 %v flow limit %v", ctx.ClientIP(), tenantInfo.Qps)
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
