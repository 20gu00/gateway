package tcp_server

import (
	"context"
	"fmt"
	"net"
	"runtime"
)

//封装的net.TCPListener
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	return tc, nil
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "tcp_proxy context value " + k.name
}

type conn struct {
	server     *TcpServer
	cancelCtx  context.CancelFunc
	rwc        net.Conn
	remoteAddr string
}

func (c *conn) close() {
	c.rwc.Close()
}

func (c *conn) serve(ctx context.Context) {
	defer func() {
		//指定recover处理链接提供服务的错误的最大次数
		if err := recover(); err != nil && err != ErrAbortHandler {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("tcp: panic serving %v: %v\n%s", c.remoteAddr, err, buf)
		}
		c.close() //关闭连接
	}()
	//远程的客户端的地址
	c.remoteAddr = c.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, LocalAddrContextKey, c.rwc.LocalAddr())
	if c.server.Handler == nil {
		panic("handler empty")
	}
	c.server.Handler.ServeTCP(ctx, c.rwc) //tcp
}
