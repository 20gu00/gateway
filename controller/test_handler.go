package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func TestHandler(c *gin.Context) {
	//c.String(200, "ok")
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "pong",
	})
}
