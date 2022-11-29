package controller

import "github.com/gin-gonic/gin"

func TestHandler(c *gin.Context) {
	c.String(200, "ok")
}
