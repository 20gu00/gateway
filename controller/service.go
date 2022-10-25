package controller

import (
	"github.com/gin-gonic/gin"
)

type ServiceController struct{}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/list", service.ServiceList)
	group.GET("/delete", service.ServiceDelete)
	group.GET("/detail", service.ServiceDetail)
	group.GET("/stat", service.ServiceStat)
	group.POST("/add_http", service.ServiceAddHttp)
	group.POST("/update_http", service.ServiceUpdateHttp)
	group.POST("/add_tcp", service.ServiceAddTcp)
	group.POST("/update_tcp", service.ServiceUpdateTcp)
	group.POST("/add_grpc", service.ServiceAddGrpc)
	group.POST("/update_grpc", service.ServiceUpdateGrpc)
}
