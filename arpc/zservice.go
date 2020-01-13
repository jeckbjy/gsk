package arpc

import (
	"errors"
)

var (
	ErrNotSupport      = errors.New("not support")
	ErrNoCodec         = errors.New("no codec")
	ErrNoHandler       = errors.New("no handler")
	ErrInvalidID       = errors.New("invalid id")
	ErrInvalidHandler  = errors.New("invalid handler")
	ErrInvalidResponse = errors.New("invalid response")
	ErrInvalidFuture   = errors.New("invalid future")
	ErrTimeout         = errors.New("timeout")
	ErrNotFoundID      = errors.New("not found id")
)

type Server interface {
	Init(opts ...Option) error
	Start() error
	Stop() error
}

// 发送的消息可以是Packet,也可以是结构体指针,
// 如果是Packet,则需要自己确保填充ID,Method等信息
// 如果是结构体指针,底层会自动创建Packet,并填充ID,Method,Name等信息
//
// Send函数:
// 	发送消息,不关系返回结果
//
// Call函数: RPC调用:分为三种形式
// 1:异步调用,rsp为回调函数
//	a:函数原型为:func(rsp *Response) error 或者func(ctx Context, rsp *Response) error
// 2:同步调用
//	a:rsp为需要返回的消息结构体指针,例如&EchoRsp{},底层会自动创建Feature
// 3:多个连续调用:限制要求,任意一次调用出错则全部失败,包括业务逻辑返回值
//	例如: 需要连续请求A,B,C三个协议,但三个协议都返回后才能继续执行
//	可以使用下面方法实现:
// 	a:外部手动创建一个Future名为f
//	b:分别请求A,B,C协议,并使用f作为参数,底层会自动为Future调用Add
//	c:调用f.Wait()方法,可以同步也可以异步
//  需要特别注意:如果外部创建Future,则必须自己手动托管Wait调用
type Client interface {
	Init(opts ...Option) error
	Send(service string, req interface{}, opts ...MiscOption) error
	Call(service string, req interface{}, rsp interface{}, opts ...MiscOption) error
}

// Future 用于异步RPC调用时,阻塞当前调用
// 需要能够支持同时多个,其中任意一次调用失败则全失败
// 调用Done时如果err不为nil,则会立刻唤醒Wait,并返回错误
type Future interface {
	Add()
	Done(err error)
	Wait() error
}
