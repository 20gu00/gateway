package tcp_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrServerClosed     = errors.New("tcp: Server closed")
	ErrAbortHandler     = errors.New("tcp: abort TCPHandler")
	ServerContextKey    = &contextKey{"tcp-server"}
	LocalAddrContextKey = &contextKey{"local-addr"}
)

type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (oc *onceCloseListener) close() {
	oc.closeErr = oc.Listener.Close()
}

type TCPHandler interface {
	ServeTCP(ctx context.Context, conn net.Conn)
}

type TcpServer struct {
	Addr    string
	Handler TCPHandler
	err     error
	BaseCtx context.Context

	WriteTimeout     time.Duration //写buf数据的超时时间
	ReadTimeout      time.Duration //读取buf数据的超时时间
	KeepAliveTimeout time.Duration //长连接超时时间

	mu         sync.Mutex
	inShutdown int32
	doneChan   chan struct{}
	l          *onceCloseListener
}

func (s *TcpServer) shuttingDown() bool {
	//判断服务是否关闭,原子操作看这个值是否等于0即关闭
	return atomic.LoadInt32(&s.inShutdown) != 0
}

func (srv *TcpServer) ListenAndServe() error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	addr := srv.Addr
	if addr == "" {
		return errors.New("need addr")
	}
	ln, err := net.Listen("tcp", addr) //建立链接,监听该地址
	if err != nil {
		return err
	}

	//提供服务(需要一个连接listner)
	return srv.Serve(tcpKeepAliveListener{ //将这个监听地址设置到长连接中
		ln.(*net.TCPListener)})
}

func (srv *TcpServer) Close() error {
	atomic.StoreInt32(&srv.inShutdown, 1)
	close(srv.doneChan) //关闭channel
	srv.l.Close()       //执行listener关闭
	return nil
}

func (srv *TcpServer) Serve(l net.Listener) error {
	srv.l = &onceCloseListener{Listener: l} //只执行一次关闭
	defer srv.l.Close()                     //关闭listener
	if srv.BaseCtx == nil {
		srv.BaseCtx = context.Background()
	}
	baseCtx := srv.BaseCtx
	//将tcp server放进context
	ctx := context.WithValue(baseCtx, ServerContextKey, srv)
	for {
		rw, e := l.Accept() //不断读取客户端发送过来的连接
		if e != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			fmt.Printf("accept fail, err: %v\n", e)
			continue
		}

		//根据获取的connection建立新的自定义的connection,
		c := srv.newConn(rw)
		go c.serve(ctx) //根据这个连接提供服务
	}
	return nil
}

func (srv *TcpServer) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc, //下游的连接,客户端
	}
	// 设置参数
	if d := c.server.ReadTimeout; d != 0 {
		c.rwc.SetReadDeadline(time.Now().Add(d))
	}
	if d := c.server.WriteTimeout; d != 0 {
		c.rwc.SetWriteDeadline(time.Now().Add(d))
	}
	if d := c.server.KeepAliveTimeout; d != 0 {
		if tcpConn, ok := c.rwc.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(d)
		}
	}
	return c
}

func (s *TcpServer) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}

func ListenAndServe(addr string, handler TCPHandler) error {
	server := &TcpServer{Addr: addr, Handler: handler, doneChan: make(chan struct{})}
	return server.ListenAndServe()
}
