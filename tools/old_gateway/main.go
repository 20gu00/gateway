package main

import (
	"flag"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/dao"
	"github.com/20gu00/gateway/httprouter"
	"github.com/20gu00/gateway/router"
	//"github.com/20gu00/gateway/tcp_proxy_router"
	"github.com/20gu00/gateway/tcprouter"
	"os"
	"os/signal"
	"syscall"
)

var (
	kind   = flag.String("kind", "", "输入服务类型 proxy or market")
	config = flag.String("config", "", "输入配置文件路径 ./conf/dev/")
)

func main() {
	flag.Parse()

	if *kind == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *kind == "market" {
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
			tcprouter.TcpServerRun() //tcp
		}()

		//接收系统信号,优雅关闭
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) //中止,强制终止
		<-quit

		//关闭可以依次
		tcprouter.TcpServerStop()
		httprouter.HttpServerStop()
		httprouter.HttpsServerStop()
	}
}
