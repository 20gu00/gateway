package common

import (
	"github.com/20gu00/gateway/common/lib"
	"github.com/garyburd/redigo/redis"
)

//里边调用操作redis的函数
func RedisConfPipeline(pip ...func(c redis.Conn)) error {
	//设置redis的连接,不用写定,而是从配置文件中获取
	c, err := lib.RedisConnFactory("default") //list.default
	if err != nil {
		return err
	}

	defer c.Close()
	for _, f := range pip {
		f(c)
	}
	c.Flush() //冲刷刷新,将数据输出到redis服务器
	return nil
}

func RedisConfDo(commandName string, args ...interface{}) (interface{}, error) {
	c, err := lib.RedisConnFactory("default") //redis链接
	if err != nil {
		return nil, err
	}

	defer c.Close()
	return c.Do(commandName, args...)
}
