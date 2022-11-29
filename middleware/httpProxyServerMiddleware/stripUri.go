package httpProxyServerMiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"strings"
)

//stripuri功能,和urlrewrite做好冲突
//访问网关127.0.0.1:9090/test/ab->期望的下游(上游服务器)的地址(实际的服务地址)127.0.0.1:9900/ab,stripuri清空多余的前缀
func HttpStripUriMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			common.Logger.Infof("未能从上下文中获取该服务详细信息")
			ctx.Abort()
			return
		}

		serviceDetail := serviceInterface.(*model.ServiceDetail)
		//前缀接入方式
		if serviceDetail.Http.RuleType == common.HTTPRuleTypePrefixURL && serviceDetail.Http.NeedStrip_uri == 1 {
			ctx.Request.URL.Path = strings.Replace(ctx.Request.URL.Path, serviceDetail.Http.Rule, "", 1) //替换一次
		}

		ctx.Next()
	}
}
