package httpmiddleware

import (
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/20gu00/gateway/reverse_proxy"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//http反向代理
func HttpReverseProxyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("未能从上下文中获取该服务详细信息"))
			ctx.Abort()
			return
		}

		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//负载均衡器
		lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			middleware.ResponseError(ctx, 2002, err)
			ctx.Abort()
			return
		}

		//连接池
		trans, err := dao.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			middleware.ResponseError(ctx, 2003, err)
			ctx.Abort()
			return
		}

		//反向代理服务器
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(ctx, lb, trans)
		proxy.ServeHTTP(ctx.Writer, ctx.Request) //提供http服务
		ctx.Abort()
		return
	}
}
