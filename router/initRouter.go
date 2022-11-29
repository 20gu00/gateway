package router

import (
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitRouter() *gin.Engine {
	//InitRouter()

	r := gin.Default()
	r.Use(middleware.LogMiddleware())

	//metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	//swaggo docs /swagger/index.html 暂时不需要swagger
	//r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))

	TestRouter(r)
	SetupRouter(r)
	return r
}
