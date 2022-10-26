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

//白名单
func HttpWhiteListMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("未能从上下文中获取该服务详细信息"))
			ctx.Abort()
			return
		}

		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		whiteIplist := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteIplist = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}

		//如果白名单为空也就是不设置限制
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whiteIplist) > 0 {
			if !common.InStringSlice(whiteIplist, ctx.ClientIP()) {
				middleware.ResponseError(ctx, 3001, errors.New(fmt.Sprintf("%s 请求的客户端ip不在ip的白名单列表中", ctx.ClientIP())))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
