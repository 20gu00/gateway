package httpmiddleware

import (
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

//http的请求header头的转换 add edit delete
func HttpHeaderTransferMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("未能从上下文中获取该服务详细信息"))
			ctx.Abort()
			return
		}

		//add h1 k1
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		for _, headerTransferItem := range strings.Split(serviceDetail.HTTPRule.HeaderTransfor, ",") {
			items := strings.Split(headerTransferItem, " ")
			if len(items) != 3 {
				continue
			}

			if items[0] == "add" || items[0] == "edit" {
				ctx.Request.Header.Set(items[1], items[2]) //设置header头,添加和修改共用set
			}

			if items[0] == "del" {
				ctx.Request.Header.Del(items[1])
			}

		}
		ctx.Next()
	}
}
