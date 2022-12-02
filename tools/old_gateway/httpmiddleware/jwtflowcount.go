package httpmiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//基于jwt统计流量
func HttpJwtFlowCountMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantInterface, ok := ctx.Get("app") //获取租户的信息
		if !ok {
			ctx.Next()
			return
		}

		tenantInfo := tenantInterface.(*dao.Tenant)
		tenantCounter, err := common.FlowCounterHandler.GetCounter(common.FlowAppPrefix + tenantInfo.AppID)
		if err != nil {
			middleware.ResponseError(ctx, 2002, err)
			ctx.Abort()
			return
		}

		tenantCounter.Increase()
		//0 不设置就是无限制
		if tenantInfo.Qpd > 0 && tenantCounter.TotalCount > tenantInfo.Qpd {
			middleware.ResponseError(ctx, 2003, errors.New(fmt.Sprintf("租户日请求量达到阈值,限流 limit:%v current:%v", tenantInfo.Qpd, tenantCounter.TotalCount)))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
