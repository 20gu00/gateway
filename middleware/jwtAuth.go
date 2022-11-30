package middleware

import (
	"fmt"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func JwtAuthTokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//Authorization: Bearer token
		//username和password编码
		token := strings.ReplaceAll(ctx.GetHeader("Authorization"), "Bearer ", "") //返回s的副本,new替换old->token
		u := new(model.Admin)
		row := dao.DB.Where("token = ?", token).First(u).RowsAffected
		fmt.Println(row)
		if row != 1 {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": 2000,
				"msg":  "当前token错误",
			})
			ctx.Abort() //ctx没有往后传,后边的请求操作也就是白,重新登录
			return
		}
		ctx.Set("UserId", u.ID)
		ctx.Next()
	}
}

//从Header头中拿token字段

//func AuthMiddleware() func(c *gin.Context) { //gin.HandlerFunc
//	return func(c *gin.Context) {
//		token := c.Request.Header.Get("token") //headers设置(不是params)
//		var u model.Admin
//		// 如果没有当前用户
//		row := dao.DB.Where("token = ?", token).First(&u).RowsAffected
//		if row != 1 {
//			c.JSON(403, gin.H{
//				"msg": "当前token错误",
//			})
//			c.Abort()
//			return
//
//		}
//		c.Set("UserId", u.ID)
//		c.Next()
//	}
//}
