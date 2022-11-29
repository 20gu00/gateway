package common

import (
	"golang.org/x/time/rate"
	"sync"
)

var FlowLimiterHandler *FlowLimiter

//基于time/rate实现限流器(限速)(流量限速,高并发三大设计,缓存限速降级)
//单例化限流器
type FlowLimiter struct {
	FlowLmiterMap   map[string]*FlowLimiterItem
	FlowLmiterSlice []*FlowLimiterItem
	Locker          sync.RWMutex
}

type FlowLimiterItem struct {
	ServiceName string
	Limter      *rate.Limiter
}

func NewFlowLimiter() *FlowLimiter {
	return &FlowLimiter{
		FlowLmiterMap:   map[string]*FlowLimiterItem{},
		FlowLmiterSlice: []*FlowLimiterItem{},
		Locker:          sync.RWMutex{},
	}
}

func init() {
	FlowLimiterHandler = NewFlowLimiter()
}

//获取flowlimiter,每个服务都有一个flowlimiter
func (counter *FlowLimiter) GetLimiter(serverName string, qps float64) (*rate.Limiter, error) {
	for _, item := range counter.FlowLmiterSlice {
		if item.ServiceName == serverName {
			return item.Limter, nil
		}
	}

	//rate新建一个limiter
	//漏桶限流,进入的速率和流出的速率,最大容纳数,如果流出小于产生(进来的请求),那么桶有可能会被装满,所以限流
	newLimiter := rate.NewLimiter(rate.Limit(qps), int(qps*3)) //每秒产生的token数目(请求数),最大的token数目
	item := &FlowLimiterItem{
		ServiceName: serverName,
		Limter:      newLimiter,
	}
	//将服务的flowlimiter写入
	counter.FlowLmiterSlice = append(counter.FlowLmiterSlice, item)
	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.FlowLmiterMap[serverName] = item
	return newLimiter, nil
}
