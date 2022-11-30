package middleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SessionAuthMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		if adminInfo, ok := session.Get(common.SessionKey).(string); !ok || adminInfo == "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": 9,
				"msg":  "未能正常获取用户的session信息(用户未登录)",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
