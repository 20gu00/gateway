package common

import (
	"fmt"
	"github.com/20gu00/gateway/common/lib"
	"github.com/garyburd/redigo/redis"
	"sync/atomic"
	"time"
)

//流量统计功能,代理使用,分布式流量统计,直接处理redis数据
type RedisFlowCountService struct {
	AppID       string        //流量统计器租户id
	Interval    time.Duration //统计间隔,刷新频率
	QPS         int64         //每秒请求量
	Unix        int64
	TickerCount int64 //滴答,计时器
	TotalCount  int64 //总数
}

//创建服务的流量统计的服务
func NewRedisFlowCountService(appID string, interval time.Duration) *RedisFlowCountService {
	reqCounter := &RedisFlowCountService{
		AppID:    appID,
		Interval: interval,
		QPS:      0,
		Unix:     0,
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()

		ticker := time.NewTicker(interval) //会根据时间间隔创建一个ticker,Ticker是一个周期触发定时的计时器，它会按照一个时间间隔往channel发送系统当前时间，而channel的接收者可以以固定的时间间隔从channel中读取事件。
		//不断循环
		for {
			<-ticker.C                                               //从周期性的计时器的channel中读取,非缓存通道,阻塞,监听事件
			tickerCount := atomic.LoadInt64(&reqCounter.TickerCount) //(同步,原子操作,实现数据的正确)获取数据,原子加载&reqCounter.TickerCount这个内存地址
			atomic.StoreInt64(&reqCounter.TickerCount, 0)            //重置数据,将val存储到&reqCounter.TickerCount

			currentTime := time.Now()
			dayKey := reqCounter.GetDayKey(currentTime)
			hourKey := reqCounter.GetHourKey(currentTime)
			//将数据存入redis,key value
			if err := RedisConfPipeline(func(c redis.Conn) {
				//设置增加和超时时间
				//写入命令,将数据发送给客户端(返回)
				c.Send("INCRBY", dayKey, tickerCount) //写入key vaule,滴答计时器的滴答数目
				c.Send("EXPIRE", dayKey, 86400*2)     //设置key超时时间
				c.Send("INCRBY", hourKey, tickerCount)
				c.Send("EXPIRE", hourKey, 86400*2)
			}); err != nil {
				fmt.Println("RedisConfPipeline,操作redis的操作器报错", err)
				continue
			}

			//从redis中拿出数据,key value
			totalCount, err := reqCounter.GetDayData(currentTime)
			if err != nil {
				fmt.Println("reqCounter请求计数器获取当天数据失败", err)
				continue
			}

			nowUnix := time.Now().Unix() //秒
			if reqCounter.Unix == 0 {
				reqCounter.Unix = time.Now().Unix()
				continue
			}

			//这个时间间隔的数目
			//计数器的值是当前的请求总数-刷新建个时长前的请求总数
			tickerCount = totalCount - reqCounter.TotalCount
			if nowUnix > reqCounter.Unix {
				reqCounter.TotalCount = totalCount
				reqCounter.QPS = tickerCount / (nowUnix - reqCounter.Unix)
				reqCounter.Unix = time.Now().Unix()
			}
		}
	}()
	return reqCounter
}

func (o *RedisFlowCountService) GetDayKey(t time.Time) string {
	//默认时区是中国
	//time.LoadLocation("Asia/Shanghai")
	dayStr := t.In(lib.TimeLocation).Format("20060102")
	//组装rediskey
	//日流量前缀+天的时间的字符串+服务流量统计前缀+服务名称
	return fmt.Sprintf("%s_%s_%s", RedisFlowDayKey, dayStr, o.AppID) //日流量前缀
}

func (o *RedisFlowCountService) GetHourKey(t time.Time) string {
	hourStr := t.In(lib.TimeLocation).Format("2006010215")
	//小时流量前缀+日的时间的字符串+服务流量统计前缀+服务名称
	return fmt.Sprintf("%s_%s_%s", RedisFlowHourKey, hourStr, o.AppID)
}

func (o *RedisFlowCountService) GetHourData(t time.Time) (int64, error) {
	//将命令转换为64为整数
	return redis.Int64(RedisConfDo("GET", o.GetHourKey(t))) //获取redis中该key的value
}

func (o *RedisFlowCountService) GetDayData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", o.GetDayKey(t)))
}

//原子增加
func (o *RedisFlowCountService) Increase() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		atomic.AddInt64(&o.TickerCount, 1)
	}()
}
