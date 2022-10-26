package main

import (
	"context"
	"fmt"
	"github.com/20gu00/gateway/tcp_server"
	"log"
	"net"
)

var (
	addr = ":7002" //本地
)

type tcpHandler struct {
}

func (t *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func main() {
	//tcp服务器测试
	log.Println("Starting tcpserver at " + addr)
	tcpServ := tcp_server.TcpServer{
		Addr:    addr,
		Handler: &tcpHandler{},
	}
	fmt.Println("Starting tcp_server at " + addr)
	tcpServ.ListenAndServe()

	//代理测试
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:6001", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()

	//redis服务器测试
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:6379", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()

	//http服务器测试:
	//缺点对请求的管控不足,比如我们用来做baidu代理,因为无法更改请求host,所以很轻易把我们拒绝
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:2003", "40")
	////rb.Add("www.baidu.com:80", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("tcp_proxy start at:" + addr)
	//tcpServ.ListenAndServe()

	//websocket服务器测试:缺点对请求的管控不足
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:2003", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()

	//http2服务器测试:缺点对请求的管控不足
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:3003", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()
}
