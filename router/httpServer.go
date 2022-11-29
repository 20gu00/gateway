package router

import (
	"context"
	"github.com/20gu00/gateway/common"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"
)

//http服务,管理后台的服务,不是直接使用gin.Engine来run,而是用go的http.Server
var (
	HttpServerHandler *http.Server
)

func HttpServerRun() {
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
	r := InitRouter()
	HttpServerHandler = &http.Server{
		Handler: r, //gin.Engine
		Addr:    viper.GetString("http.address"),
		//time.Duration单位纳秒
		ReadTimeout:    time.Duration(viper.GetInt("http.read_timeout") * 1000000000),
		WriteTimeout:   time.Duration(viper.GetInt("http.write_timeout") * 1000000000),
		MaxHeaderBytes: 1 << viper.GetInt("http.max_header_bytes"),
	}

	common.Logger.Infof("后台管理 HttpServerRun:%s\n", viper.GetString("http.address"))
	if err := HttpServerHandler.ListenAndServe(); err != nil {
		common.Logger.Infof("后台管理 HttpServerRun:  %s\n error:%v\n", viper.GetString("http.address"), err)
	}
}

func HttpServerStop() {
	//超时即关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpServerHandler.Shutdown(ctx); err != nil {
		common.Logger.Infof("后台管理 HttpServer Stop error:%v\n", err)
	}
	common.Logger.Infof("后台管理 HttpServer Stop\n")
}
