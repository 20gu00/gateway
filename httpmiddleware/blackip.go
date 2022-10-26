package httpmiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

//ip黑名单,白名单优先级高于黑名单,白名单匹配放行,黑名单匹配不放行
func HttpBlackListMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("未能从上下文中获取服务详细信息"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail) //断言->servicedetail
		whileIpList := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			//写入白名单配置
			whileIpList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}

		blackIpList := []string{}
		if serviceDetail.AccessControl.BlackList != "" {
			//写入黑名单配置
			blackIpList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}
		//开启验证并且白名单为空 白名单优先级高于黑名单
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whileIpList) == 0 && len(blackIpList) > 0 {
			if common.InStringSlice(blackIpList, ctx.ClientIP()) { //解析X-Real-IP和X-Forwarded-For,返回真是客户端ip
				middleware.ResponseError(ctx, 3001, errors.New(fmt.Sprintf("%s处于黑名单列表", ctx.ClientIP())))
				ctx.Abort()
				return
			}
		}
		ctx.Next() //不是黑名单ip就继续向下处理
	}
}
