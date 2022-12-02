package httprouter

import (
	"context"
	"github.com/20gu00/gateway/common/lib"
	"github.com/20gu00/gateway/middleware"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	HttpSrvHandler  *http.Server
	HttpsSrvHandler *http.Server
)

//http和https的启动和停止的方法

func HttpServerRun() {
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode")) //生产环境
	r := InitRouter(middleware.RecoveryMiddleware(),        //gin的框架实例,包含了中间件等,提供服务
		middleware.RequestLog())
	//初始化http server配置
	HttpSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.http.addr"), //网关对外提供http服务的端口
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.http.read_timeout")) * time.Second,  //读取超时时长(代理)
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.http.write_timeout")) * time.Second, //写入超时时长(代理)
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.http.max_header_bytes")),                //二进制位长度,最大的header头大小
	}
	log.Printf(" [INFO] http代理服务器运行 %s\n", lib.GetStringConf("proxy.http.addr"))
	//ListenAndServe
	if err := HttpSrvHandler.ListenAndServe(); err != nil && err != http.ErrServerClosed { //非关闭服务的err
		log.Fatalf(" [ERROR] http代理服务器运行 %s err:%v\n", lib.GetStringConf("proxy.http.addr"), err)
	}
}

func HttpsServerRun() {
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode"))
	r := InitRouter(middleware.RecoveryMiddleware(),
		middleware.RequestLog())
	HttpsSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.https.addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.https.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.https.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.https.max_header_bytes")),
	}
	log.Printf(" [INFO] https代理服务器运行 %s\n", lib.GetStringConf("proxy.https.addr"))
	//ListenAndServeTLS
	if err := HttpsSrvHandler.ListenAndServeTLS("./cert/server.crt", "./cert/server.key"); err != nil && err != http.ErrServerClosed {
		log.Fatalf(" [ERROR] https代理服务器运行 %s err:%v\n", lib.GetStringConf("proxy.https.addr"), err)
	}
}

func HttpServerStop() {
	//定义超时关闭的context,函数间传递,控制这个goroutine的生命周期,服务关闭的最大时间
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //超时时间,截止时间=超时时间+当前时间
	defer cancel()                                                           //操作完成就释放
	if err := HttpSrvHandler.Shutdown(ctx); err != nil {
		log.Printf(" [ERROR] http代理服务器停止 err:%v\n", err)
	}
	log.Printf(" [INFO] http代理服务器停止 %v stopped\n", lib.GetStringConf("proxy.http.addr"))
}

func HttpsServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpsSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] https代理服务器停止 err:%v\n", err)
	}
	log.Printf(" [INFO] https代理服务器停止 %v stopped\n", lib.GetStringConf("proxy.https.addr"))
}
