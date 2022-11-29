package common

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"os"
)

func RedisConfPipeline(pip ...func(c redis.Conn)) error {
	workDir, err := os.Getwd()
	if err != nil {
		Logger.Infof("获取工作目录失败")
		return err
	}

	//读取配置文件
	viper.SetConfigName("redis")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf") //可多个
	if err := viper.ReadInConfig(); err != nil {
		Logger.Infof("redis配置文件读取失败", err.Error())
		return err
	}

	c, err := redis.Dial("tcp", viper.GetString("redis.proxy_list"))
	if err != nil {
		Logger.Infof("连接redis失败")
	}

	defer c.Close()
	res, err := c.Do("ping")
	fmt.Println(res)

	//每一个函数都包含了众多的redis操作
	for _, f := range pip {
		f(c)
	}
	c.Flush() //冲刷刷新,将数据输出到redis服务器
	return nil
}

func RedisConfDo(commandName string, args ...interface{}) (interface{}, error) {
	workDir, err := os.Getwd()
	if err != nil {
		Logger.Infof("获取工作目录失败")
		return nil, err
	}

	//读取配置文件
	viper.SetConfigName("redis")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf") //可多个
	if err := viper.ReadInConfig(); err != nil {
		Logger.Infof("redis配置文件读取失败", err.Error())
		return nil, err
	}

	c, err := redis.Dial("tcp", viper.GetString("redis.proxy_list"))
	if err != nil {
		Logger.Infof("连接redis失败")
	}

	defer c.Close()
	return c.Do(commandName, args...)
}
