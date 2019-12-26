package arpc

import (
	"context"
	"sync/atomic"
)

var defaultRouter atomic.Value

func DefaultRouter() Router {
	return defaultRouter.Load().(Router)
}

func SetDefaultRouter(r Router) {
	defaultRouter.Store(r)
}

func Register(callback interface{}, opts ...RegisterOption) error {
	return DefaultRouter().Register(callback, opts...)
}

// Router 消息处理路由,需要支持服务器响应和客户端RPC响应两种形式:
// 1:服务端消息响应,这个比较简单,一般服务器启动时注册好消息回调就可以了,
// 	 可以通过消息ID,消息名,或者调用方法名查找到对应的消息回调
// Router 消息回调路由
// 1:静态注册:根据消息ID,消息名或者调用方法名找到对应的消息回调
// 2:RPC消息回调,根据SeqID查找消息回调
//
// 静态回调函数原型:
// 原型1: func(ctx Context) error
// 原型2: func(ctx Context, req *Request) error
// 原型3: func(ctx Context, req *Request, rsp *Response) error
// 原型1需要手动解析,原型2,3需要反射,性能会有额外消耗
// 原型3:系统会自动发送消息,但是如何确定消息的ID呢?
// RPC回调原型:
// rpc调用不需要Request信息,只通过SeqID查询
// 原型1: func(rsp *Response) error
// 原型2: func(ctx Context, rsp *Response) error
//
// 获取Endpoints信息
// TODO:Router的实现还是笼统庞杂了,也许应该明确区分出来到底是使用哪种通信协议
type Router interface {
	// 注册消息回调
	Register(callback interface{}, opts ...RegisterOption) error
	// 注册RPC回调
	RegisterRPC(req Packet) error
	// Find 根据消息包,查询消息回调,可能是服务端响应,也可能是客户端RPC响应
	Find(pkg Packet) (Handler, error)
	Close() error
}

type RegisterOption func(o *RegisterOptions)
type RegisterOptions struct {
	Context context.Context
	ID      uint   // 消息ID
	Name    string // 消息名或方法名
	Method  string // 调用方法名
}

// 重试回调函数,返回error则终止重试
type RetryFunc func(req Packet, count int) error

// rpc注册时需要用到的数据
type CallInfo struct {
	RetryCB  RetryFunc
	Future   Future
	Response interface{}
}
