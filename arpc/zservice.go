package arpc

import (
	"context"
	"errors"
	"time"

	"github.com/jeckbjy/gsk/registry"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/selector"
)

var (
	ErrNoHandler      = errors.New("no handler")
	ErrInvalidHandler = errors.New("invalid handler")
	ErrNotSupport     = errors.New("not support")
)

const (
	DefaultCallTTL = time.Second
)

type Server interface {
	Init(opts ...Option) error
	Start() error
	Stop() error
}

type Client interface {
	Init(opts ...Option) error
	// 发送消息到指定service
	Send(service string, req interface{}, opts ...CallOption) error
	// RPC调用:分为三种形式
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
	Call(service string, req interface{}, rsp interface{}, opts ...CallOption) error
}

// Future 用于异步RPC调用时,阻塞当前调用
// 需要能够支持同时多个,其中任意一次调用失败则全失败
// 调用Done时如果err不为nil,则会立刻唤醒Wait,并返回错误
type Future interface {
	Add()
	Done(err error)
	Wait() error
}

type Option func(o *Options)
type Options struct {
	Context   context.Context
	Registry  registry.Registry
	Tran      anet.Tran
	Name      string            // 服务名
	Id        string            // 服务ID
	Version   string            // 服务版本
	Address   string            // Listen使用
	Advertise string            // 注册服务使用
	Selector  selector.Selector // client load balance
	Proxy     string            // 代理服务名,空代表不使用代理
}

type CallOption func(o *CallOptions)
type CallOptions struct {
	selector.Options
	ID     int           // 消息ID
	Future Future        // 异步调用
	Retry  int           // 重试次数
	TTL    time.Duration // 超时时间
}
