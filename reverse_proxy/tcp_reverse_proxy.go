package reverse_proxy

import (
	"context"
	"github.com/20gu00/gateway/reverse_proxy/load_balance"
	"github.com/20gu00/gateway/tcpmiddleware"
	"io"
	"log"
	"net"
	"time"
)

//新建个tc代理服务器,支持负载均衡
//往负载均衡器中添加服务器的地址,在这个代理过程中,服务器属于上游,但从数据源的过程来看,也可以是下游
func NewTcpLoadBalanceReverseProxy(c *tcpmiddleware.TcpSliceRouterContext, lb load_balance.LoadBalance) *TcpReverseProxy {
	return func() *TcpReverseProxy {
		nextAddr, err := lb.Get("") //获取下游的地址
		if err != nil {
			log.Fatal("get next addr fail")
		}
		return &TcpReverseProxy{
			ctx:             c.Ctx,
			Addr:            nextAddr,
			KeepAlivePeriod: time.Second,
			DialTimeout:     time.Second,
		}
	}()
}

//TCP反向代理
type TcpReverseProxy struct {
	ctx                  context.Context //单次请求单独设置
	Addr                 string
	KeepAlivePeriod      time.Duration //设置
	DialTimeout          time.Duration //设置超时时间
	DialContext          func(ctx context.Context, network, address string) (net.Conn, error)
	OnDialError          func(src net.Conn, dstDialErr error)
	ProxyProtocolVersion int
}

func (dp *TcpReverseProxy) dialTimeout() time.Duration {
	if dp.DialTimeout > 0 {
		return dp.DialTimeout
	}
	return 10 * time.Second
}

var defaultDialer = new(net.Dialer)

func (dp *TcpReverseProxy) dialContext() func(ctx context.Context, network, address string) (net.Conn, error) {
	if dp.DialContext != nil {
		return dp.DialContext
	}
	return (&net.Dialer{
		Timeout:   dp.DialTimeout,     //连接超时
		KeepAlive: dp.KeepAlivePeriod, //设置连接的检测时长
	}).DialContext
}

func (dp *TcpReverseProxy) keepAlivePeriod() time.Duration {
	if dp.KeepAlivePeriod != 0 {
		return dp.KeepAlivePeriod
	}
	return time.Minute
}

//传入上游 conn，在这里完成下游连接与数据交换 src-访问->dst
func (dp *TcpReverseProxy) ServeTCP(ctx context.Context, src net.Conn) {
	//设置连接超时,设置context超时
	var cancel context.CancelFunc
	if dp.DialTimeout >= 0 {
		ctx, cancel = context.WithTimeout(ctx, dp.dialTimeout())
	}
	dst, err := dp.dialContext()(ctx, "tcp", dp.Addr) //下游地址
	if cancel != nil {
		cancel()
	}
	if err != nil {
		dp.onDialError()(src, err)
		return
	}

	defer func() { go dst.Close() }() //关闭下游连接(客户端)

	//设置dst连接
	if ka := dp.keepAlivePeriod(); ka > 0 {
		if c, ok := dst.(*net.TCPConn); ok { //tcp连接
			c.SetKeepAlive(true)     //设置tcp连接为长连接
			c.SetKeepAlivePeriod(ka) //保持连接的周期
		}
	}
	errc := make(chan error, 1) //缓存1
	//上游下游数据交换
	go dp.proxyCopy(errc, src, dst)
	go dp.proxyCopy(errc, dst, src)
	<-errc
}

func (dp *TcpReverseProxy) onDialError() func(src net.Conn, dstDialErr error) {
	if dp.OnDialError != nil {
		return dp.OnDialError
	}
	return func(src net.Conn, dstDialErr error) {
		log.Printf("tcpproxy: for incoming conn %v, error dialing %q: %v", src.RemoteAddr().String(), dp.Addr, dstDialErr)
		src.Close()
	}
}

func (dp *TcpReverseProxy) proxyCopy(errc chan<- error, dst, src net.Conn) {
	_, err := io.Copy(dst, src) //数据拷贝,两个socker传输数据
	errc <- err
}
