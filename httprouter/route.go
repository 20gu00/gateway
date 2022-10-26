package httprouter

import (
	"github.com/20gu00/gateway/controller"
	"github.com/20gu00/gateway/httpmiddleware"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
)

//初始化中间件
func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	//router := gin.Default()  会输出内容到终端
	router := gin.New()
	router.Use(middlewares...)
	//router.Handle
	router.GET("/ping", func(c *gin.Context) { //请求路径 匹配处理逻辑函数
		c.JSON(200, gin.H{ //json序列化,放入resp,map[string]interface
			"message": "pong",
		})
	})

	//路由组
	oauth := router.Group("/oauth")
	oauth.Use(middleware.TranslationMiddleware())
	{
		controller.AuthRegister(oauth)
	}

	//全局中间件,一系列handlefunc(洋葱)
	router.Use(
		httpmiddleware.HttpAccessModeMiddleware(),     //域名还是前缀
		httpmiddleware.HttpFlowCountMiddleware(),      //流量统计
		httpmiddleware.HttpFlowLimitMiddleware(),      //限流
		httpmiddleware.HttpJwtAuthTokenMiddleware(),   //jwt认证,服务访问的权限认证
		httpmiddleware.HttpJwtFlowCountMiddleware(),   //jwt流量统计
		httpmiddleware.HttpJwtFlowLimitMiddleware(),   //jwt限流
		httpmiddleware.HttpWhiteListMiddleware(),      //白名单
		httpmiddleware.HttpBlackListMiddleware(),      //黑名单
		httpmiddleware.HttpHeaderTransferMiddleware(), //header头转换
		httpmiddleware.HttpStripUriMiddleware(),       //strip_uri
		httpmiddleware.HttpUrlRewriteMiddleware(),     //urlrewrite
		httpmiddleware.HttpReverseProxyMiddleware())   //http反向代理

	return router //*gin.Engine
}
