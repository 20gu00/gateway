package tcprouter

import (
	"context"
	"fmt"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/reverse_proxy"
	"github.com/20gu00/gateway/tcp_server"
	"github.com/20gu00/gateway/tcpmiddleware"
	"log"
	"net"
)

//网关的tcp代理服务(负载均衡),接入方式port

var tcpServerList = []*tcp_server.TcpServer{}

type tcpHandler struct {
}

func (t *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func TcpServerRun() {
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}

			//构建路由及设置中间件
			router := tcpmiddleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcpmiddleware.TCPFlowCountMiddleware(),
				tcpmiddleware.TCPFlowLimitMiddleware(),
				tcpmiddleware.TCPWhiteListMiddleware(),
				tcpmiddleware.TCPBlackListMiddleware(),
			)

			//构建回调handler
			routerHandler := tcpmiddleware.NewTcpSliceRouterHandler(
				func(c *tcpmiddleware.TcpSliceRouterContext) tcp_server.TCPHandler {
					return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, rb)
				}, router)

			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)
			tcpServer := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerHandler,
				BaseCtx: baseCtx,
			}
			tcpServerList = append(tcpServerList, tcpServer)
			log.Printf(" [INFO] tcp_proxy_run %v\n", addr)
			if err := tcpServer.ListenAndServe(); err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf(" [INFO] tcp_proxy_run %v err:%v\n", addr, err)
			}
		}(tempItem)
	}
}

func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
}
