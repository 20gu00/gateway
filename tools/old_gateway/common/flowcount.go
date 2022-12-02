package common

import (
	"sync"
	"time"
)

var FlowCounterHandler *FlowCounter

//单例化flow counter流量统计器,和loadbalancer类似,每个服务一个流量统计器
type FlowCounter struct {
	RedisFlowCountMap   map[string]*RedisFlowCountService
	RedisFlowCountSlice []*RedisFlowCountService
	Locker              sync.RWMutex
}

func NewFlowCounter() *FlowCounter {
	return &FlowCounter{
		RedisFlowCountMap:   map[string]*RedisFlowCountService{},
		RedisFlowCountSlice: []*RedisFlowCountService{},
		Locker:              sync.RWMutex{},
	}
}

func init() {
	FlowCounterHandler = NewFlowCounter()
}

//获取一个统计器
func (counter *FlowCounter) GetCounter(serverName string) (*RedisFlowCountService, error) {
	for _, item := range counter.RedisFlowCountSlice { //全部统计器的统计业务
		if item.AppID == serverName {
			return item, nil //返回该被统计流量的服务的流量信息
		}
	}

	//实际的创建统计器的逻辑
	//如果没有统计该服务的流量信息,就创建流量统计功能,服务流量统计器的appid就是流量统计前缀+服务名
	newCounter := NewRedisFlowCountService(serverName, 1*time.Second)
	//将新的流量统计器信息添加到流量统计器的切片和map中
	counter.RedisFlowCountSlice = append(counter.RedisFlowCountSlice, newCounter)
	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.RedisFlowCountMap[serverName] = newCounter
	return newCounter, nil
}
