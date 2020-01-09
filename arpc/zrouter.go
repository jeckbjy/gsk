package arpc

import (
	"sync/atomic"

	"github.com/jeckbjy/gsk/anet"
)

var gContextFactory ContextFactory
var gRouter atomic.Value

func SetContextFactory(fn ContextFactory) {
	gContextFactory = fn
}

func NewContext() Context {
	return gContextFactory()
}

func SetRouter(r Router) {
	gRouter.Store(r)
}

func GetRouter() Router {
	return gRouter.Load().(Router)
}

// Handler 消息回调处理函数
type Handler func(ctx Context) error

// Middleware 中间件,next可能为nil,即找不到handler
type Middleware func(next Handler) Handler

type ContextFactory func() Context

// Handler 上下文
type Context interface {
	Init(conn anet.Conn, msg Packet)    // 初始化
	Free()                              // 释放,可用于pool回收
	Get(key string) (interface{}, bool) // 根据key获取数据
	Set(key string, val interface{})    // 根据key设置数据
	Data() interface{}                  // 自定义数据
	SetData(v interface{})              // 设置数据
	Error() error                       // 错误信息,比如Timeout
	SetError(err error)                 // 设置错误
	Conn() anet.Conn                    // 原始Socket
	Message() Packet                    // 消息
	Send(msg interface{}) error         // 发送消息,不关心返回结果
}

// 消息路由
// 消息类型上分为两种:
// 一:客户端请求消息,Ack为false,回调函数通常静态注册
// 二:服务器应答消息,Ack为true, 回调函数通常由调用Call时注册
// 应用场景上分两种:
// 一:普通的消息注册与查询
// 二:代理请求,通常只需要根据规则转发,通常使用全局静态函数,但是需要上下文参数
//
// 消息处理支持中间件,可用于异常处理,消息统计过滤,全局代理也可以使用中间件进行处理
type Router interface {
	Use(middleware ...Middleware)
	Handle(ctx Context) error
	Register(cb interface{}, opts ...MiscOption) error
}
