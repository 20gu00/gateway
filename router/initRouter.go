package router

import (
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitRouter() *gin.Engine {
	//InitRouter()

	r := gin.Default()
	r.Use(middleware.LogMiddleware(), sessions.Sessions("mysession", dao.Store)) //如果使用github.com/gin-gonic/contrib/session需要调用session这个中间,也就是设置了DefaultKey否则会报错

	//metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	//swaggo docs /swagger/index.html 暂时不需要swagger
	//r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))  //gs "github.com/swaggo/gin-swagger"  "github.com/swago/gin-swagger/swaggerFiles"

	TestRouter(r)
	SetupRouter(r)
	return r
}
