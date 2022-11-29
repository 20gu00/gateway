package router

import (
	"github.com/20gu00/gateway/controller"
	"github.com/gin-gonic/gin"
)

func TestRouter(e *gin.Engine) {
	//v1:=e.Group("/api/v1")
	e.GET("/ping", controller.TestHandler)
}
