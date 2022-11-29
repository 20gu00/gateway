package dao

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	//"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/spf13/viper"
	"os"
)

var Store sessions.Store

func InitSessionRedis() {
	workDir, err := os.Getwd()
	if err != nil {
		common.Logger.Infof("获取工作目录失败")
		return
	}

	//读取配置文件
	viper.SetConfigName("redis")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf")
	if err := viper.ReadInConfig(); err != nil {
		common.Logger.Infof("redis配置文件读取失败", err.Error())
		return
	}

	//密钥是成对定义的，以允许密钥旋转，但常见的情况是设置一个单一的认证密钥和可选的加密密钥。
	store, err := sessions.NewRedisStore(viper.GetInt("redis.max_idle"), "tcp", viper.GetString("redis.proxy_list"), viper.GetString("redis.password"), []byte("secret"))
	if err != nil {
		fmt.Println(err)
		return
	}
	//使用redis作为session存储引擎
	//store, err := redis.NewStore(viper.GetInt("redis.max_idle"), "tcp", viper.GetString("redis.proxy_list"), viper.GetString("redis.password"), []byte("secret"))
	Store = store
}
