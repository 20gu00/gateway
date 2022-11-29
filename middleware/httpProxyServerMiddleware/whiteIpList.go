package httpProxyServerMiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"strings"
)

//白名单
func HttpWhiteListMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			common.Logger.Infof("未能从上下文中获取该服务详细信息")
			ctx.Abort()
			return
		}

		serviceDetail := serviceInterface.(*model.ServiceDetail)
		whiteIplist := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteIplist = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}

		//如果白名单为空也就是不设置限制
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whiteIplist) > 0 {

			//发起请求的客户端的ip不属于服务定义的白名单
			if !common.StrInSlice(whiteIplist, ctx.ClientIP()) {
				common.Logger.Infof(fmt.Sprintf("%s 请求的客户端ip不在ip的白名单列表中", ctx.ClientIP()))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
