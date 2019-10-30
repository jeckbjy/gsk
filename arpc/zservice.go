package arpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jeckbjy/gsk/arpc/selector"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/registry"
)

var (
	ErrTimeout    = errors.New("rpc timeout")
	ErrNoHandler  = errors.New("no handler")
	ErrNoNode     = errors.New("no node found")
	ErrNoConn     = errors.New("no conn found")
	ErrRetryStop  = errors.New("retry stop")
	ErrNotSupport = errors.New("not support")
)

// 全双工的服务,包括服务器和客户端
type Service interface {
	Server
	Client
	Options() *Options
	Init(opts ...Option) error
}

type Server interface {
	Register(srv interface{})
	Start() error
	Stop() error
	Wait() error
	Run() error
}

type Client interface {
	Send(service string, msg interface{}) error
	Call(service string, req interface{}, rsp interface{}, opts ...CallOption) error
}

// Future 用于异步RPC调用时,阻塞当前调用
// 需要能够支持同时多个,其中任意一次调用失败则全失败
type Future interface {
	Add(delta int) error
	Done() error
	Fail(err error) error
	Wait() error
}

const (
	DisableServer   = 0x01
	DisableClient   = 0x02
	DisableRegistry = 0x04
)

type Option func(*Options)
type Options struct {
	ServerOptions
	ClientOptions
	Flags    int
	Tran     anet.Tran
	Registry registry.Registry
	Context  context.Context
	Router   Router
}

func (o *Options) HasFlag(mask int) bool {
	return (o.Flags & mask) == mask
}

type ServerOptions struct {
	Name        string         // 服务名
	Id          string         // 服务ID
	Version     string         // 服务版本
	Address     string         // Listen使用
	Advertise   string         // 注册服务使用
	BeforeStart []func() error //
	AfterStart  []func() error
	BeforeStop  []func() error
	AfterStop   []func() error
}

func (o *ServerOptions) FullId() string {
	return fmt.Sprintf("%s-%s", o.Name, o.Id)
}

type ClientOptions struct {
	Selector  selector.Selector
	RPCRouter RPCRouter // 用于注册RPC回调,通常不需要配置
	Services  []string  // 需要监听的服务,更多配置?
	Proxy     string    // 代理服务名
}

type CallOption func(o *CallOptions)
type CallOptions struct {
	ID     int           // 消息ID
	Name   string        // 消息名
	Method string        // 调用函数名
	Future Future        // 异步调用
	Retry  int           // 重试次数
	TTL    time.Duration // 超时时间
}
