package httpProxyServerMiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)

//url重写
func HttpUrlRewriteMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			common.Logger.Infof("未能从上下文中获取该服务详细信息")
			ctx.Abort()
			return
		}

		serviceDetail := serviceInterface.(*model.ServiceDetail)
		//匹配前 匹配后,匹配前 匹配后  多个重写规则
		//127.0.0.1:9090:/test/ab->127.0.0.1:9090:/test/ba
		for _, urlRewriteItem := range strings.Split(serviceDetail.Http.UrlRewrite, ",") {
			items := strings.Split(urlRewriteItem, " ")
			if len(items) != 2 {
				continue
			}

			regexp, err := regexp.Compile(items[0]) //正则解析,返回regexp
			if err != nil {
				continue //正则匹配错误直接下一个循环,没必要return(ResponseError中间件)
			}

			//^/test/ab(.*) /test/ba$1  捕获分组
			replacePath := regexp.ReplaceAll([]byte(ctx.Request.URL.Path), []byte(items[1]))
			ctx.Request.URL.Path = string(replacePath)
		}
		ctx.Next()
	}
}
