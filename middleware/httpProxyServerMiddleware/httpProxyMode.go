package httpProxyServerMiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model/manager"
	"github.com/gin-gonic/gin"
)

//处理http请求接入方式,前缀(路径)和域名
func HttpProxyModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := manager.ServiceManagerHandler.HttpAccessMode(c)
		if err != nil {
			common.Logger.Infof("判断http请求的接入方式和获取该服务的详情失败", err.Error())
			fmt.Println(err)
			c.Abort() //取消,不在向下传递
			return
		}

		c.Set("service", service) //传递servicedetail
		c.Next()                  //传递给子context
	}
}
