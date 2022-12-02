package main

import (
	"flag"
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/initdo"
	"github.com/20gu00/gateway/router"
	"github.com/20gu00/gateway/router/httpProxyServer"
	"os"
	"time"
)

var (
	kind = flag.String("kind", "admin", "输入要开启的服务器类型 proxy or admin or all")
	c    = 1
)

func main() {
	flag.Parse()
	initdo.InitDo()
	if *kind == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *kind == "admin" {
		initdo.Admin <- c
	} else if *kind == "proxy" {
		initdo.Proxy <- c
	} else if *kind == "all" {
		initdo.All <- c
	} else {
		fmt.Printf("输入参数不正确 proxy or admin or all")
		os.Exit(1)
	}

	go func() {
		stop := common.SignalHandler()
		select {
		case <-stop:
			if *kind == "admin" {
				router.HttpServerStop()
				<-time.NewTimer(10 * time.Second).C
				os.Exit(0)
			} else if *kind == "proxy" {
				httpProxyServer.HttpProxyServerStop()
				httpProxyServer.HttpsProxyServerStop()
				<-time.NewTimer(10 * time.Second).C
				//<-time.After(10*time.Minute)
				os.Exit(0)
			} else {
				router.HttpServerStop()
				httpProxyServer.HttpProxyServerStop()
				httpProxyServer.HttpsProxyServerStop()
				<-time.NewTimer(10 * time.Second).C
				os.Exit(0)
			}
		}
	}()

	for {
		select {
		case <-initdo.Admin:
			go func() {
				router.HttpServerRun()
			}()
		case <-initdo.Proxy:
			go func() {
				httpProxyServer.HttpProxyServerRun()
			}()
			go func() {
				httpProxyServer.HttpsProxyServerRun()
			}()
		case <-initdo.All:
			go func() {
				router.HttpServerRun()
			}()
			go func() {
				httpProxyServer.HttpProxyServerRun()
			}()
			go func() {
				httpProxyServer.HttpsProxyServerRun()
			}()
		}
	}

	//监听退出信号,模拟优雅关闭(关闭handler 数据库连接等)
	//quit := make(chan os.Signal)
	//signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT) //监听退出信号(sigint类比ctrl-c,sigterm正常退出
	//<-quit                                               //阻塞
	//如果某个服务未开启,这里使用关闭,业务虽然不影响,但是逻辑不好
	//router.HttpServerStop()
	//httpProxyServer.HttpProxyServerStop()
	//httpProxyServer.HttpsProxyServerStop()
	//tcpProxyServer.TcpProxyServerStop()
}

//func main() {
//	initdo.InitDo()
//	r := router.InitRouter()
//	r.Run(":8000")
//}
