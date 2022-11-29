package httpProxyServerMiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
)

//基于jwt统计流量(租户访问某个服务的流量统计)
func HttpJwtFlowCountMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantInterface, ok := ctx.Get("app") //获取租户的信息
		if !ok {
			ctx.Next()
			return
		}

		tenantInfo := tenantInterface.(*model.Tenant)
		tenantCounter, err := common.FlowCounterHandler.GetCounter(common.FlowTenantPrefix + tenantInfo.AppId)
		if err != nil {
			common.Logger.Infof("创建统计器失败(用途统计某个租户访问摸个服务的流量统计)")
			ctx.Abort()
			return
		}

		tenantCounter.Increase()

		//0 不设置就是无限制
		if tenantInfo.Qpd > 0 && tenantCounter.TotalCount > int64(tenantInfo.Qpd) {
			common.Logger.Infof("租户日请求量达到阈值,限流 limit:%v current:%v", tenantInfo.Qpd, tenantCounter.TotalCount)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
