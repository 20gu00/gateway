package middleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/20gu00/gateway/model/manager"
	"github.com/gin-gonic/gin"
	"strings"
)

//jwt验证,token认证
func HttpJwtAuthTokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverInterface, ok := ctx.Get("service")
		if !ok {
			common.Logger.Infof("未能从上下文中获取服务详细信息")
			ctx.Abort()
			return
		}

		serviceDetail := serverInterface.(*model.ServiceDetail)

		//Header
		//Authorization: Bearer token(也是Header头的key/vale信息
		//username和password编码
		token := strings.ReplaceAll(ctx.GetHeader("Authorization"), "Bearer ", "") //返回s的副本,new替换old->token
		appMatched := false                                                        //初始化未匹配
		if token != "" {
			claims, err := common.JwtDecode(token) //*jwt.StandardClaims,标准token字段(访问群体,过期时间,id,发行时间,发行机构)
			if err != nil {
				common.Logger.Infof("解码token失败")
				ctx.Abort()
				return
			}

			appList := manager.AppManagerHandler.GetTenantList()
			for _, appInfo := range appList {
				if appInfo.AppId == claims.Issuer { //租户id和token发行人匹配
					ctx.Set("app", appInfo) //设置租户信息进context
					appMatched = true       //成功匹配
					break
				}
			}
		}

		//服务开启了权限认证并且没有匹配到租户信息,无权访问
		if serviceDetail.AccessControl.OpenAuth == 1 && !appMatched {
			common.Logger.Infof("未匹配到租户")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
