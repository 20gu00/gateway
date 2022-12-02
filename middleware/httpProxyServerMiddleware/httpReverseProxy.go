package httpProxyServerMiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/loadBalance"
	"github.com/20gu00/gateway/common/reverseProxyServer"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
)

//http反向代理
func HttpReverseProxyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverInterface, ok := ctx.Get("service")
		if !ok {
			common.Logger.Infof("未能从上下文中获取该服务详细信息")
			ctx.Abort()
			return
		}

		serviceDetail := serverInterface.(*model.ServiceDetail)

		//http基于负载均衡器和连接池的代理

		//负载均衡器
		lb, err := loadBalance.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			common.Logger.Infof("创建负载均衡器失败")
			ctx.Abort()
			return
		}

		//连接池(连接实际的工作负载)
		trans, err := loadBalance.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			common.Logger.Infof("获取后端实际工作负载失败")
			ctx.Abort()
			return
		}

		//反向代理服务器
		proxy := reverseProxyServer.NewLoadBalanceReverseProxy(ctx, lb, trans)
		proxy.ServeHTTP(ctx.Writer, ctx.Request) //提供http服务(也是http服务暴露出去)
		ctx.Abort()
		return
	}
}
