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
	router := gin.New() //gin的router
	router.Use(middlewares...)
	//router.Handle
	router.GET("/ping", func(c *gin.Context) { //请求路径 匹配处理逻辑函数
		c.JSON(200, gin.H{ //json序列化,放入resp,map[string]interface
			"message": "pong",
		})
	})

	//创建路由组,再制定这个路由组使用的中间件,注册的子路由也使用这些中间件
	//先匹配
	oauth := router.Group("/auth")
	oauth.Use(middleware.TranslationMiddleware()) //多语言转换中间件
	{
		controller.AuthRegister(oauth)
	}

	//全局中间件,一系列handlefunc(洋葱)
	//oauth := router.Group("/")
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
