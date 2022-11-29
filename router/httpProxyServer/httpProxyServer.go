package httpProxyServer

import (
	"context"
	"github.com/20gu00/gateway/common"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"time"
)

//http的代理服务器
var (
	HttpPorxyServerHandler  *http.Server //指针
	HttpsPorxyServerHandler *http.Server
)

func HttpProxyServerRun() {
	workDir, err := os.Getwd()
	if err != nil {
		common.Logger.Infof("获取工作目录失败")
		return
	}

	viper.SetConfigName("general")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf")
	if err := viper.ReadInConfig(); err != nil {
		common.Logger.Infof("general配置文件读取失败", err.Error())
		return
	}

	//initdo.InitDo()
	r := InitProxyRouter()
	HttpPorxyServerHandler = &http.Server{
		Handler:        r, //gin.Engine
		Addr:           viper.GetString("proxy.http.addr"),
		ReadTimeout:    time.Duration(viper.GetInt("proxy.http.read_timeout")),
		WriteTimeout:   time.Duration(viper.GetInt("proxy.http.write_timeout")),
		MaxHeaderBytes: 1 << viper.GetInt("proxy.http.max_header_bytes"),
	}

	common.Logger.Infof("http代理服务器 HttpProxyServerRun:%s\n", viper.GetString("proxy.http.addr"))
	if err := HttpPorxyServerHandler.ListenAndServe(); err != nil {
		common.Logger.Infof("http代理服务器 HttpProxyServerRun:%s\n error:%v\n", viper.GetString("proxy.http.addr"), err)
	}
}

//关闭自然是打印出日志而不是终端
func HttpProxyServerStop() {
	//超时即关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpPorxyServerHandler.Shutdown(ctx); err != nil {
		log.Fatalf("http代理服务器 HttpProxyServer Stop error:%v\n", err)
	}
	log.Printf("http代理服务器 HttpProxyServer Stop\n")
}

func HttpsProxyServerRun() {
	workDir, err := os.Getwd()
	if err != nil {
		common.Logger.Infof("获取工作目录失败")
		return
	}

	viper.SetConfigName("general")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf")
	if err := viper.ReadInConfig(); err != nil {
		common.Logger.Infof("general配置文件读取失败", err.Error())
		return
	}

	//initdo.InitDo()
	r := InitProxyRouter()
	HttpsPorxyServerHandler = &http.Server{
		Handler:        r, //gin.Engine
		Addr:           viper.GetString("proxy.https.addr"),
		ReadTimeout:    time.Duration(viper.GetInt("proxy.https.read_timeout")),
		WriteTimeout:   time.Duration(viper.GetInt("proxy.https.write_timeout")),
		MaxHeaderBytes: 1 << viper.GetInt("proxy.https.max_header_bytes"),
	}

	common.Logger.Infof("https 代理服务器 HttpsProxyServerRun:%s\n", viper.GetString("http.address"))
	if err := HttpsPorxyServerHandler.ListenAndServe(); err != nil {
		common.Logger.Infof("https 代理服务器 HttpsProxyServerRun:%s\n error:%v\n", viper.GetString("http.address"), err)
	}
}

func HttpsProxyServerStop() {
	//超时即关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpsPorxyServerHandler.Shutdown(ctx); err != nil {
		common.Logger.Infof("https 代理服务器 HttpsProxyServer Stop error:%v\n", err)
	}
	common.Logger.Infof("https 代理服务器 HttpsProxyServer Stop\n")
}
