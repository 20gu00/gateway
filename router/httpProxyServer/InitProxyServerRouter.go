package httpProxyServer

import (
	"github.com/20gu00/gateway/middleware"
	"github.com/20gu00/gateway/middleware/httpProxyServerMiddleware"
	"github.com/20gu00/gateway/router"
	"github.com/gin-gonic/gin"
)

func InitProxyRouter() *gin.Engine {
	r := gin.Default()
	//Group("/")  顺序
	r.Use(
		middleware.LogMiddleware(),                               //gin的处理请求的日志也使用logrus
		httpProxyServerMiddleware.HttpProxyModeMiddleware(),      //请求接入方式
		httpProxyServerMiddleware.HttpFlowCountMiddleware(),      //流量统计
		httpProxyServerMiddleware.HttpFlowLimitMiddleware(),      //限流
		httpProxyServerMiddleware.HttpJwtAuthTokenMiddleware(),   //基于jwt的认证(可以实现一定的权限认证)
		httpProxyServerMiddleware.HttpJwtFlowCountMiddleware(),   //jwt的流量统计
		httpProxyServerMiddleware.HttpJwtFlowLimitMiddleware(),   //租户的流量统计
		httpProxyServerMiddleware.HttpWhiteListMiddleware(),      //白名单
		httpProxyServerMiddleware.HttpBlackListMiddleware(),      //黑名单
		httpProxyServerMiddleware.HttpHeaderTransferMiddleware(), //header头转换
		httpProxyServerMiddleware.HttpStripUriMiddleware(),       //strip_uri
		httpProxyServerMiddleware.HttpUrlRewriteMiddleware(),     //url rewrite
	)
	router.TestRouter(r)
	return r
}
