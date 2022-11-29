package tcpProxyServer

import (
	"context"
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/common/loadBalance"
	"github.com/20gu00/gateway/model"
	"github.com/20gu00/gateway/model/manager"
	"github.com/20gu00/gateway/tools/tcpProxy"
	tcpProxyServerMiddleware2 "github.com/20gu00/gateway/tools/tcpProxy/tcpProxyServerMiddleware"
	"github.com/20gu00/gateway/tools/tcpProxy/tcpServer"
	"log"
	"net"
)

//网关的tcp代理服务(负载均衡),接入方式port

var tcpServerList = []*tcpServer.TcpServer{}

func (t *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

type tcpHandler struct{}

func TcpProxyServerRun() {
	//获取tcp服务列表
	serviceList := manager.ServiceManagerHandler.GetTcpServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		//启动tcp server
		go func(serviceDetail *model.ServiceDetail) {
			//拿到端口,启动tcp server需要(代理)
			addr := fmt.Sprintf(":%d", serviceDetail.Tcp.Port)

			//设置负载均衡器
			rb, err := loadBalance.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				//直接关闭整个程序
				log.Fatalf("[INFO] 创建处理tcp服务的负载均衡器 %v err:%v\n", addr, err)
				return
			}

			//构建路由及设置中间件,非gin的router(tcp的server)
			router := tcpProxyServerMiddleware2.NewTcpSliceRouter()
			router.Group("/").Use( //即相应的ip端口下流量全部进入这里
				tcpProxyServerMiddleware2.TCPFlowCountMiddleware(),
				tcpProxyServerMiddleware2.TCPFlowLimitMiddleware(),
				tcpProxyServerMiddleware2.TCPWhiteListMiddleware(),
				tcpProxyServerMiddleware2.TCPBlackListMiddleware(),
			)

			//使用路由构建回调方法handler,业务逻辑处理
			routerHandler := tcpProxyServerMiddleware2.NewTcpSliceRouterHandler(
				func(c *tcpProxyServerMiddleware2.TcpSliceRouterContext) tcpServer.TCPHandler {
					return tcpProxy.NewTcpLoadBalanceReverseProxy(c, rb)
				}, router)

			//基于自定义的context创建context,传递serivce,值是servicedetail
			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)

			//提供tcp server(这里不是代理)
			//创建tcp server
			tcpServerObj := &tcpServer.TcpServer{
				Addr:    addr,
				Handler: routerHandler, //使用路由构建的回调函数
				BaseCtx: baseCtx,
			}
			//将tcp server放进tcp server的列表中
			tcpServerList = append(tcpServerList, tcpServerObj)
			common.Logger.Infof("[INFO] tcp代理服务启动 %v\n", addr)

			//启动tcp server
			if err := tcpServerObj.ListenAndServe(); err != nil && err != tcpServer.ErrServerClosed {
				common.Logger.Infof("[INFO] tcp代理服务启动 %v err:%v\n", addr, err)
			}

		}(tempItem)
	}
}

func TcpProxyServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf("[INFO] tcp代理服务停止 %v stopped\n", tcpServer.Addr)
	}
}
