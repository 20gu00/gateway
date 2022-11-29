package router

import (
	"github.com/20gu00/gateway/controller"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SetupRouter(e *gin.Engine) {
	//admin
	admin := e.Group("/admin")
	//如果使用github.com/gin-gonic/contrib/session需要调用session这个中间,也就是设置了DefaultKey否则会报错
	admin.Use(sessions.Sessions("mysession", dao.Store))
	admin.POST("/login", controller.AdminLoginHandler)
	admin.POST("/register", controller.AdminRegisterHandler)
	admin.Use(middleware.AuthMiddleware())
	admin.POST("/changepwd", controller.ChangePwdHandler)
	admin.POST("/logout", controller.LoginOutHandler)

	//service
	service := e.Group("/service")
	service.Use(middleware.AuthMiddleware(), sessions.Sessions("mysession", dao.Store))
	service.GET("/list", controller.ServiceListHandler)
	service.DELETE("/delete", controller.ServiceDeleteHandler)
	service.GET("/detail", controller.ServiceDetailHandler)
	service.GET("/stat", controller.ServiceStatHandler)
	service.POST("/add_http", controller.ServiceAddHttpHandler)
	service.PUT("/update_http", controller.ServiceUpdateHttpHandler)
	service.POST("/add_tcp", controller.ServiceAddTcpHandler)
	service.PUT("/update_tcp", controller.ServiceUpdateTcpHandler)

	//tenant
	tenant := e.Group("tenant")
	tenant.Use(middleware.AuthMiddleware(), sessions.Sessions("mysession", dao.Store))
	tenant.GET("/list", controller.TenantListHandler)
	tenant.GET("/detail", controller.TenantDetailHandler)
	tenant.GET("/stat", controller.TenantStatHandler)
	tenant.GET("/delete", controller.TenantDeleteHandler)
	tenant.POST("/add", controller.TenantAddHandler)
	tenant.POST("/update", controller.TenantUpdateHandler)

	//dash
	dash := e.Group("dash")
	dash.Use(middleware.AuthMiddleware(), sessions.Sessions("mysession", dao.Store))
	dash.GET("/panel", service.PanelDataHandler)
	dash.GET("/flow_stat", service.FlowStatHandler)
}
