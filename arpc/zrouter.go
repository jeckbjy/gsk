package arpc

import (
	"context"
	"time"
)

// Router 消息处理路由,需要支持服务器响应和客户端RPC响应两种形式:
// 1:服务端消息响应,这个比较简单,一般服务器启动时注册好消息回调就可以了,
// 	 可以通过消息ID,消息名,或者调用方法名查找到对应的消息回调
type Router interface {
	// Register:可用于注册消息回调,也可以用于注册消息解析,或者注册RPC回调
	// 1:注册函数,即消息回调,同时也会自动反射解析需要的消息名
	// 3:注册服务,反射所有符合条件的函数,用REGISTER_SRV标识
	// 支持的函数原型有:
	// 原型1:func(ctx IContext)
	// 原型2:func(ctx IContext, req *Request) error
	// 原型3:func(ctx IContext, req *Request, rsp *Response) error
	// 原型1调用不需要反射,效率最高,但是则需要额外手动注册消息解析
	// 原型2,3使用了反射调用,应该会有额外调用消耗,可以通过方法名或者消息名找到消息回调
	Register(srv interface{}) error
	//// 注册消息,需要消息原型,消息名(通过反射获得),消息ID(可选)
	//RegisterMsg(msg interface{}) error
	// Find 根据消息包,查询消息回调,可能是服务端响应,也可能是客户端RPC响应
	Find(pkg Packet) (Handler, error)
	// 通过名字查询ID,返回0表示没有找到
	//FindID(name string) int
}

// RPCRouter 用于管理RPC回调函数
type RPCRouter interface {
	// 注册RPC,
	// 同步调用rsp是一个消息指针,必须提供一个用于阻塞用的Future才有意义
	// 异步调用是一个回调函数, 原型是func(ctx IContext, rsp *XXResponse)
	Register(rsp interface{}, o *RegisterRPCOptions) error
	// 查询回调
	Find(seqId string) (Handler, error)
}

//
type IIDProvider interface {
	FindID(name string) uint
}

// 重试函数,返回error则终止重试
type RetryFunc func(count int) (SeqID string, TTL time.Duration, err error)

// 注册
type RegisterRPCOptions struct {
	Context  context.Context
	SeqID    string
	Future   Future        // 完成通知,同步调用才需要,异步函数执行完才会通知
	TTL      time.Duration // 超时时间,必须设置
	RetryMax int           // 最大重试次数
	RetryCB  RetryFunc     // 重试回调
}
