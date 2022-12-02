package httpmiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//基于jwt的限流
func HttpJwtFlowLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantInterface, ok := ctx.Get("app")
		if !ok {
			ctx.Next()
			return
		}

		tenantInfo := tenantInterface.(*dao.Tenant)
		if tenantInfo.Qps > 0 {
			clientLimiter, err := common.FlowLimiterHandler.GetLimiter(common.FlowAppPrefix+tenantInfo.AppID+"_"+ctx.ClientIP(), float64(tenantInfo.Qps))
			if err != nil {
				middleware.ResponseError(ctx, 5001, err)
				ctx.Abort()
				return
			}

			if !clientLimiter.Allow() {
				middleware.ResponseError(ctx, 5002, errors.New(fmt.Sprintf("租户限流 %v flow limit %v", ctx.ClientIP(), tenantInfo.Qps)))
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
