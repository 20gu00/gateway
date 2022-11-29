package tcpProxyServerMiddleware

import (
	"context"
	"github.com/20gu00/gateway/tools/tcpProxy/tcpServer"
	"math"
	"net"
)

const abortIndex int8 = math.MaxInt8 / 2 //最多 63 个中间件

type TcpHandlerFunc func(*TcpSliceRouterContext)

//使用方法数组,实现tcp支持的中间件模式
//路由器->/匹配(全部tcp流量都匹配进来)->方法数组->方法一->方法二->tcp方向代理
// router 结构体
type TcpSliceRouter struct {
	groups []*TcpSliceGroup
}

// group 结构体
type TcpSliceGroup struct {
	*TcpSliceRouter //指针
	path            string
	handlers        []TcpHandlerFunc
}

// router上下文
type TcpSliceRouterContext struct {
	conn net.Conn
	Ctx  context.Context
	*TcpSliceGroup
	index int8
}

func newTcpSliceRouterContext(conn net.Conn, r *TcpSliceRouter, ctx context.Context) *TcpSliceRouterContext {
	newTcpSliceGroup := &TcpSliceGroup{}
	*newTcpSliceGroup = *r.groups[0] //浅拷贝数组指针,只会使用第一个分组
	c := &TcpSliceRouterContext{conn: conn, TcpSliceGroup: newTcpSliceGroup, Ctx: ctx}
	c.Reset()
	return c
}

func (c *TcpSliceRouterContext) Get(key interface{}) interface{} {
	return c.Ctx.Value(key)
}

func (c *TcpSliceRouterContext) Set(key, val interface{}) {
	c.Ctx = context.WithValue(c.Ctx, key, val)
}

//封装了路由器和回调方法,满足tcp serverde要求
type TcpSliceRouterHandler struct {
	coreFunc func(*TcpSliceRouterContext) tcpServer.TCPHandler
	router   *TcpSliceRouter
}

func (w *TcpSliceRouterHandler) ServeTCP(ctx context.Context, conn net.Conn) {
	c := newTcpSliceRouterContext(conn, w.router, ctx)
	c.handlers = append(c.handlers, func(c *TcpSliceRouterContext) {
		w.coreFunc(c).ServeTCP(ctx, conn)
	})
	c.Reset()
	c.Next()
}

func NewTcpSliceRouterHandler(coreFunc func(*TcpSliceRouterContext) tcpServer.TCPHandler, router *TcpSliceRouter) *TcpSliceRouterHandler {
	return &TcpSliceRouterHandler{
		coreFunc: coreFunc,
		router:   router,
	}
}

// 构造 router
func NewTcpSliceRouter() *TcpSliceRouter {
	return &TcpSliceRouter{}
}

// 创建 Group
func (g *TcpSliceRouter) Group(path string) *TcpSliceGroup {
	if path != "/" {
		panic("path要求是/")
	}
	return &TcpSliceGroup{
		TcpSliceRouter: g,
		path:           path,
	}
}

// 构造回调方法
func (g *TcpSliceGroup) Use(middlewares ...TcpHandlerFunc) *TcpSliceGroup {
	g.handlers = append(g.handlers, middlewares...) //将中间件添加到路由组的回调函数集合中,方法数组,多个
	existsFlag := false
	for _, oldGroup := range g.TcpSliceRouter.groups {
		if oldGroup == g {
			existsFlag = true
		}
	}
	if !existsFlag {
		g.TcpSliceRouter.groups = append(g.TcpSliceRouter.groups, g)
	}
	return g
}

// 从最先加入中间件开始回调
func (c *TcpSliceRouterContext) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// 跳出中间件方法
func (c *TcpSliceRouterContext) Abort() {
	c.index = abortIndex
}

// 是否跳过了回调
func (c *TcpSliceRouterContext) IsAborted() bool {
	return c.index >= abortIndex
}

// 重置回调
func (c *TcpSliceRouterContext) Reset() {
	c.index = -1
}