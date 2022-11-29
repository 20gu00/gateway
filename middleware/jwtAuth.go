package middleware

import (
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() func(c *gin.Context) { //gin.HandlerFunc
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token") //headers设置(不是params)
		var u model.Admin
		// 如果没有当前用户
		row := dao.DB.Where("token = ?", token).First(&u).RowsAffected
		if row != 1 {
			c.JSON(403, gin.H{
				"msg": "当前token错误",
			})
			c.Abort()
			return

		}
		c.Set("UserId", u.ID)
		c.Next()
	}
}
