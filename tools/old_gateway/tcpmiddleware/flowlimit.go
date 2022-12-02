package tcpmiddleware

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
	"strings"
)

//客户端流量控制,服务流量控制
func TCPFlowLimitMiddleware() func(c *TcpSliceRouterContext) {
	return func(c *TcpSliceRouterContext) {
		serverInterface := c.Get("service")
		if serverInterface == nil {
			c.conn.Write([]byte("get service empty"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		if serviceDetail.AccessControl.ServiceFlowLimit != 0 {
			serviceLimiter, err := common.FlowLimiterHandler.GetLimiter(
				common.FlowServicePrefix+serviceDetail.ServiceInfo.ServiceName,
				float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				c.conn.Write([]byte(err.Error()))
				c.Abort()
				return
			}
			if !serviceLimiter.Allow() {
				c.conn.Write([]byte(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				c.Abort()
				return
			}
		}

		splits := strings.Split(c.conn.RemoteAddr().String(), ":")
		clientIP := ""
		if len(splits) == 2 {
			clientIP = splits[0]
		}
		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := common.FlowLimiterHandler.GetLimiter(
				common.FlowServicePrefix+serviceDetail.ServiceInfo.ServiceName+"_"+clientIP,
				float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				c.conn.Write([]byte(err.Error()))
				c.Abort()
				return
			}
			if !clientLimiter.Allow() {
				c.conn.Write([]byte(fmt.Sprintf("%v flow limit %v", clientIP, serviceDetail.AccessControl.ClientIPFlowLimit)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
