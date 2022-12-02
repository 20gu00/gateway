package common

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	FlowCounterHandler *FlowCounter
	TimeLocation, _    = time.LoadLocation("Asia/Shanghai")
)

//单例化flow counter流量统计器,和loadbalancer类似,每个服务一个流量统计器
type FlowCounter struct {
	RedisFlowCountMap   map[string]*FlowCountService
	RedisFlowCountSlice []*FlowCountService
	Locker              sync.RWMutex
}

func NewFlowCounter() *FlowCounter {
	return &FlowCounter{
		RedisFlowCountMap:   map[string]*FlowCountService{},
		RedisFlowCountSlice: []*FlowCountService{},
		Locker:              sync.RWMutex{},
	}
}

func init() {
	FlowCounterHandler = NewFlowCounter()
}

//获取一个统计器
func (counter *FlowCounter) GetCounter(serverName string) (*FlowCountService, error) {
	for _, item := range counter.RedisFlowCountSlice { //全部统计器的统计业务
		if item.AppID == serverName {
			return item, nil //返回该被统计流量的服务的流量信息
		}
	}

	//实际的创建统计器的逻辑
	//如果没有统计该服务的流量信息,就创建流量统计功能,服务流量统计器的appid就是流量统计前缀+服务名
	newCounter := NewFlowCountService(serverName, 1*time.Second) //设置每秒统计一次,也就是每秒向redis写入(操作)一次数据
	//将新的流量统计器信息添加到流量统计器的切片和map中
	counter.RedisFlowCountSlice = append(counter.RedisFlowCountSlice, newCounter)
	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.RedisFlowCountMap[serverName] = newCounter
	return newCounter, nil
}

type FlowCountService struct {
	AppID       string        //流量统计器租户id(非网关租户)(标示那一个流量统计器)
	Interval    time.Duration //统计间隔,刷新频率
	QPS         int64         //每秒请求量
	Unix        int64
	TickerCount int64 //滴答,计时器
	TotalCount  int64 //总数
}

//创建流量统计的服务,统计功能实现
func NewFlowCountService(appID string, interval time.Duration) *FlowCountService {
	reqCounter := &FlowCountService{
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
			<-ticker.C                                               //阻塞一定的时间                                            //从周期性的计时器的channel中读取,非缓存通道,阻塞,监听事件
			tickerCount := atomic.LoadInt64(&reqCounter.TickerCount) //(同步,原子操作,实现数据的正确)获取数据,原子加载&reqCounter.TickerCount这个内存地址
			atomic.StoreInt64(&reqCounter.TickerCount, 0)            //重置数据,将val存储到&reqCounter.TickerCount

			currentTime := time.Now()
			//redis的key(key value形式存放每日和每天的流量)
			dayKey := reqCounter.GetDayKey(currentTime)
			hourKey := reqCounter.GetHourKey(currentTime)
			//将数据存入redis,key value
			if err := RedisConfPipeline(func(c redis.Conn) {
				//设置增加和超时时间
				//写入命令,将数据发送给客户端(返回)
				c.Send("INCRBY", dayKey, tickerCount) //写入key vaule,滴答计时器的滴答数目,将 key 中储存的数字加上指定的增量值。 如果key 不存在,那么 key 的值会先被初始化为 0 ,然后再执行 INCRBY 命令
				c.Send("EXPIRE", dayKey, 86400*2)     //设置key超时时间
				c.Send("INCRBY", hourKey, tickerCount)
				c.Send("EXPIRE", hourKey, 86400*2) //两天
			}); err != nil {
				fmt.Println("RedisConfPipeline,操作redis的操作器报错", err)
				continue
			}

			//从redis中拿出数据,key value
			totalCount, err := reqCounter.GetDayData(currentTime)
			if err != nil {
				fmt.Println("reqCounter 请求计数器获取当天数据失败", err)
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

func (f *FlowCountService) GetDayKey(t time.Time) string {
	workDir, err := os.Getwd()
	if err != nil {
		Logger.Infof("获取工作目录失败")
		return ""
	}

	viper.SetConfigName("general")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf")
	if err := viper.ReadInConfig(); err != nil {
		Logger.Infof("general配置文件读取失败", err.Error())
		return ""
	}

	//12345 2006-01-02 15:04:05
	TimeLocation, _ = time.LoadLocation(viper.GetString("time.time_loc"))
	dayStr := t.In(TimeLocation).Format("20060102")
	//组装rediskey(流量统计相关的数据存放在redis中)
	//qpd 日流量前缀+天的时间的字符串+流量统计器的id
	return fmt.Sprintf("%s_%s_%s", DayFlowStatKey, dayStr, f.AppID) //日流量前缀
}

func (f *FlowCountService) GetHourKey(t time.Time) string {
	workDir, err := os.Getwd()

	if err != nil {
		Logger.Infof("获取工作目录失败")
		return ""
	}

	viper.SetConfigName("general")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir + "/conf")
	if err := viper.ReadInConfig(); err != nil {
		Logger.Infof("general配置文件读取失败", err.Error())
		return ""
	}

	//12345 2006-01-02 15:04:05
	TimeLocation, _ = time.LoadLocation(viper.GetString("time.time_loc"))
	hourStr := t.In(TimeLocation).Format("2006010215")
	//小时流量前缀+日的时间的字符串+服务流量统计前缀+服务名称
	return fmt.Sprintf("%s_%s_%s", RedisFlowHourKey, hourStr, f.AppID)
}

func (f *FlowCountService) GetHourData(t time.Time) (int64, error) {
	//将命令转换为64位整数
	return redis.Int64(RedisConfDo("GET", f.GetHourKey(t))) //获取redis中该key的value
}

func (f *FlowCountService) GetDayData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", f.GetDayKey(t)))
}

//原子增加
func (f *FlowCountService) Increase() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		atomic.AddInt64(&f.TickerCount, 1)
	}()
}
