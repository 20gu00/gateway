package main

import (
	"flag"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/grpc_proxy_router"
	"github.com/20gu00/gateway/httprouter"
	"github.com/20gu00/gateway/router"
	"github.com/20gu00/gateway/tcp_proxy_router"
	"os"
	"os/signal"
	"syscall"
)

var (
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
	config   = flag.String("config", "", "input config file like ./conf/dev/")
)

func main() {
	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *endpoint == "dashboard" {
		lib.InitModule(*config)
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		lib.InitModule(*config)
		defer lib.Destroy()
		dao.ServiceManagerHandler.LoadOnce()
		dao.AppManagerHandler.LoadOnce()

		//多个代理服务器
		go func() {
			httprouter.HttpServerRun() //http
		}()
		go func() {
			httprouter.HttpsServerRun() //https
		}()
		go func() {
			tcp_proxy_router.TcpServerRun() //tcp
		}()
		go func() {
			grpc_proxy_router.GrpcServerRun() //grpc
		}()

		//接收系统信号,优雅关闭
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) //中止,强制终止
		<-quit

		//关闭可以依次
		tcp_proxy_router.TcpServerStop()
		grpc_proxy_router.GrpcServerStop()
		httprouter.HttpServerStop()
		httprouter.HttpsServerStop()
	}
}
