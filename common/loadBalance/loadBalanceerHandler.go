package loadBalance

import (
	"fmt"
	"github.com/20gu00/gateway/common"
	"github.com/20gu00/gateway/model"
	"net"
	"net/http"
	"sync"
	"time"
)

var LoadBalancerHandler *LoadBalancer

type LoadBalancer struct {
	LoadBanlanceMap   map[string]*LoadBalancerItem
	LoadBanlanceSlice []*LoadBalancerItem
	Locker            sync.RWMutex
}

type LoadBalancerItem struct {
	LoadBanlance LoadBalance
	ServiceName  string
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		LoadBanlanceMap:   map[string]*LoadBalancerItem{},
		LoadBanlanceSlice: []*LoadBalancerItem{},
		Locker:            sync.RWMutex{},
	}
}

func init() {
	LoadBalancerHandler = NewLoadBalancer()
}

func (lbr *LoadBalancer) GetLoadBalancer(service *model.ServiceDetail) (LoadBalance, error) {
	for _, lbrItem := range lbr.LoadBanlanceSlice {
		if lbrItem.ServiceName == service.BaseInfo.ServiceName {
			return lbrItem.LoadBanlance, nil
		}
	}
	schema := "http://"
	if service.Http.NeedHttps == 1 {
		schema = "https://"
	}

	//ip:port
	if service.BaseInfo.LoadType == common.LoadTypeTCP {
		schema = ""
	}
	ipList := service.LoadBalance.GetIpList()
	weightList := service.LoadBalance.GetWeightList()
	ipConf := map[string]string{}
	for ipIndex, ipItem := range ipList {
		ipConf[ipItem] = weightList[ipIndex]
	}
	mConf, err := NewLoadBalanceCheckConf(fmt.Sprintf("%s%s", schema, "%s"), ipConf)
	if err != nil {
		return nil, err
	}
	lb := LoadBanlanceFactorWithConf(LbType(service.LoadBalance.RoundType), mConf)

	lbItem := &LoadBalancerItem{
		LoadBanlance: lb,
		ServiceName:  service.BaseInfo.ServiceName,
	}
	lbr.LoadBanlanceSlice = append(lbr.LoadBanlanceSlice, lbItem)

	lbr.Locker.Lock()
	defer lbr.Locker.Unlock()
	lbr.LoadBanlanceMap[service.BaseInfo.ServiceName] = lbItem
	return lb, nil
}

var TransportorHandler *Transportor

type Transportor struct {
	TransportMap   map[string]*TransportItem
	TransportSlice []*TransportItem
	Locker         sync.RWMutex
}

type TransportItem struct {
	Trans       *http.Transport
	ServiceName string
}

func NewTransportor() *Transportor {
	return &Transportor{
		TransportMap:   map[string]*TransportItem{},
		TransportSlice: []*TransportItem{},
		Locker:         sync.RWMutex{},
	}
}

func init() {
	TransportorHandler = NewTransportor()
}

//返回http的连接池
func (t *Transportor) GetTrans(service *model.ServiceDetail) (*http.Transport, error) {
	for _, transItem := range t.TransportSlice {
		if transItem.ServiceName == service.BaseInfo.ServiceName {
			return transItem.Trans, nil
		}
	}

	if service.LoadBalance.UpstreamConnectTimeOut == 0 {
		service.LoadBalance.UpstreamConnectTimeOut = 30
	}
	if service.LoadBalance.UpstreamMaxIdle == 0 {
		service.LoadBalance.UpstreamMaxIdle = 100
	}
	if service.LoadBalance.UpstreamIdleTimeOut == 0 {
		service.LoadBalance.UpstreamIdleTimeOut = 90
	}
	if service.LoadBalance.UpstreamHeaderTimeOut == 0 {
		service.LoadBalance.UpstreamHeaderTimeOut = 30
	}
	trans := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(service.LoadBalance.UpstreamConnectTimeOut) * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          service.LoadBalance.UpstreamMaxIdle,
		IdleConnTimeout:       time.Duration(service.LoadBalance.UpstreamIdleTimeOut) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: time.Duration(service.LoadBalance.UpstreamHeaderTimeOut) * time.Second,
	}

	//save to map and slice
	transItem := &TransportItem{
		Trans:       trans,
		ServiceName: service.BaseInfo.ServiceName,
	}
	t.TransportSlice = append(t.TransportSlice, transItem)
	t.Locker.Lock()
	defer t.Locker.Unlock()
	t.TransportMap[service.BaseInfo.ServiceName] = transItem
	return trans, nil
}
