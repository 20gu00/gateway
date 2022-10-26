package tcpmiddleware

import (
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/dao"
)

func TCPFlowCountMiddleware() func(c *TcpSliceRouterContext) {
	return func(c *TcpSliceRouterContext) {
		serverInterface := c.Get("service")
		if serverInterface == nil {
			c.conn.Write([]byte("get service empty"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//全站  服务
		totalCounter, err := common.FlowCounterHandler.GetCounter(common.FlowTotal)
		if err != nil {
			c.conn.Write([]byte(err.Error()))
			c.Abort()
			return
		}
		totalCounter.Increase() //每次请求进来就加一

		serviceCounter, err := common.FlowCounterHandler.GetCounter(common.FlowServicePrefix + serviceDetail.ServiceInfo.ServiceName)
		if err != nil {
			c.conn.Write([]byte(err.Error()))
			c.Abort()
			return
		}
		serviceCounter.Increase()
		c.Next()
	}
}
