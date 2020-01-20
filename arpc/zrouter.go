package arpc

import (
	"sync/atomic"

	"github.com/jeckbjy/gsk/anet"
)

var gContextFactory ContextFactory
var gIDProvider IDProvider
var gRouter atomic.Value

func SetContextFactory(fn ContextFactory) {
	gContextFactory = fn
}

func NewContext() Context {
	return gContextFactory()
}

// SetRouter 设置全局Router
func SetRouter(r Router) {
	gRouter.Store(r)
}

// GetRouter 获取全局Router
func GetRouter() Router {
	return gRouter.Load().(Router)
}

// Use 设置全局Middleware
func Use(middleware HandlerFunc) {
	GetRouter().Use(middleware)
}

func SetIDProvider(p IDProvider) {
	gIDProvider = p
}

func GetIDProvider() IDProvider {
	return gIDProvider
}

// Handler 消息回调处理函数
type HandlerFunc func(ctx Context) error
type HandlerChain []HandlerFunc

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
	Handler() HandlerFunc               // 获取最终要执行的回调,可能为nil
	SetHandler(h HandlerFunc)           // 设置Handler
	SetMiddleware(h HandlerChain)       // 设置Middleware
	Conn() anet.Conn                    // 原始Socket
	Message() Packet                    // 消息
	Response() Packet                   // 应答消息
	SetResponse(rsp Packet)             // 按需设置Response
	Send(msg interface{}) error         // 发送消息,不关心返回结果
	Abort(err error)                    // 手动中止调用
	Next() error                        // 调用下一个,返回错误则自动中止
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
	Use(middleware ...HandlerFunc)
	Handle(ctx Context) error
	Register(cb interface{}, opts ...MiscOption) error
}

// MessageID 用于通过反射识别消息是否提供了消息ID,从而避免通过Name映射查询ID
// 接口函数可以使用工具自动生成代码
type MessageID interface {
	MsgID() int
}

// IDProvider 用于消息名和ID一一映射
// 纯粹的RPC调用并不需要填充消息ID
// 但如果是以MsgID作为唯一标识的情况下,需要提供MsgID才能保证客户端能够查询到消息回调
// 在测试环境下,可以使用名字作为唯一标识,使用SetBindName设置绑定开关,默认false
type IDProvider interface {
	SetBindName(flag bool)
	Register(name string, id int) error
	GetID(name string) int
	GetName(id int) string
	Fill(packet Packet, msg interface{}) error
}
