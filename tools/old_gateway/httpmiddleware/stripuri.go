package httpmiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

//stripuri功能,和rulrewrite做好冲突
//访问网关127.0.0.1:9090/test/ab->期望的下游的地址(实际的服务地址)127.0.0.1:9900/ab,stripuri清空多余的前缀
func HttpStripUriMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("未能从上下文中获取该服务详细信息"))
			ctx.Abort()
			return
		}

		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		//前缀接入方式
		if serviceDetail.HTTPRule.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.HTTPRule.NeedStripUri == 1 {
			ctx.Request.URL.Path = strings.Replace(ctx.Request.URL.Path, serviceDetail.HTTPRule.Rule, "", 1) //替换一次
		}

		ctx.Next()
	}
}
