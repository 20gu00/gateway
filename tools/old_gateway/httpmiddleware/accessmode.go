package httpmiddleware

import (
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
)

//处理http接入方式的中间件,前缀(路径)和域名
//*gin.HandlerFunc作为gin路由配置参数
func HttpAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := dao.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 1001, err)
			c.Abort() //中止context不在传递
			return
		}

		c.Set("service", service) //传递servicedetail
		c.Next()                  //传递给子context
	}
}
