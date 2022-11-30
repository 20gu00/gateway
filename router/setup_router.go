package router

import (
	"github.com/20gu00/gateway/controller"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(e *gin.Engine) {
	//admin
	admin := e.Group("/admin")
	admin.POST("/register", controller.AdminRegisterHandler)
	admin.POST("/login", controller.AdminLoginHandler)
	admin.Use(middleware.JwtAuthTokenMiddleware(), middleware.SessionAuthMiddleware())
	admin.POST("/changepwd", controller.ChangePwdHandler)
	admin.POST("/logout", controller.LoginOutHandler)
	admin.POST("/info", controller.AdminInfoHandler)

	//service
	service := e.Group("/service")
	service.Use(middleware.JwtAuthTokenMiddleware(), middleware.SessionAuthMiddleware())
	service.GET("/list", controller.ServiceListHandler)
	service.DELETE("/delete", controller.ServiceDeleteHandler)
	service.GET("/detail", controller.ServiceDetailHandler)
	service.GET("/stat", controller.ServiceStatHandler)
	service.POST("/add_http", controller.ServiceAddHttpHandler)
	service.PUT("/update_http", controller.ServiceUpdateHttpHandler)
	//service.POST("/add_tcp", controller.ServiceAddTcpHandler)
	//service.PUT("/update_tcp", controller.ServiceUpdateTcpHandler)

	//tenant
	tenant := e.Group("tenant")
	tenant.Use(middleware.JwtAuthTokenMiddleware(), middleware.SessionAuthMiddleware())
	tenant.GET("/list", controller.TenantListHandler)
	tenant.GET("/detail", controller.TenantDetailHandler)
	tenant.GET("/stat", controller.TenantStatHandler)
	tenant.GET("/delete", controller.TenantDeleteHandler)
	tenant.POST("/add", controller.TenantAddHandler)
	tenant.POST("/update", controller.TenantUpdateHandler)

	//dash
	dash := e.Group("dash")
	dash.Use(middleware.JwtAuthTokenMiddleware(), middleware.SessionAuthMiddleware())
	dash.GET("/panel", controller.PanelDataHandler)
	dash.GET("/flow_stat", controller.FlowStatHandler)
}
