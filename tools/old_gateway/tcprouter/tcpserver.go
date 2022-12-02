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

type tcpHandler struct{}

func (t *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func TcpServerRun() {
	//获取tcp服务列表
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		//启动tcp server
		go func(serviceDetail *dao.ServiceDetail) {
			//拿到端口,启动tcp server需要(代理)
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)

			//设置负载均衡器
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				//直接关闭整个程序
				log.Fatalf(" [INFO] 获取tcp服务的负载均衡器 %v err:%v\n", addr, err)
				return
			}

			//构建路由及设置中间件,非gin的router
			router := tcpmiddleware.NewTcpSliceRouter()
			router.Group("/").Use( //即相应的ip端口下流量全部进入这里
				tcpmiddleware.TCPFlowCountMiddleware(),
				tcpmiddleware.TCPFlowLimitMiddleware(),
				tcpmiddleware.TCPWhiteListMiddleware(),
				tcpmiddleware.TCPBlackListMiddleware(),
			)

			//使用路由构建回调方法handler,业务逻辑处理
			routerHandler := tcpmiddleware.NewTcpSliceRouterHandler(
				func(c *tcpmiddleware.TcpSliceRouterContext) tcp_server.TCPHandler {
					return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, rb)
				}, router)

			//基于自定义的context创建context,传递serivce,值是servicedetail
			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)

			//提供tcp server(这里不是代理)
			//创建tcp server
			tcpServer := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerHandler, //使用路由构建的回调函数
				BaseCtx: baseCtx,
			}
			//将tcp server放进tcp server的列表中
			tcpServerList = append(tcpServerList, tcpServer)
			log.Printf(" [INFO] tcp代理服务启动 %v\n", addr)

			//启动tcp server
			if err := tcpServer.ListenAndServe(); err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf(" [INFO] tcp代理服务启动 %v err:%v\n", addr, err)
			}

		}(tempItem)
	}
}

func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] tcp代理停止 %v stopped\n", tcpServer.Addr)
	}
}
