package arpc

import (
	"context"
	"errors"
	"time"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc/selector"
	"github.com/jeckbjy/gsk/registry"
)

var (
	ErrTimeout        = errors.New("rpc timeout")
	ErrNoHandler      = errors.New("no handler")
	ErrInvalidHandler = errors.New("invalid handler")
	ErrNoNode         = errors.New("no node found")
	ErrNoConn         = errors.New("no conn found")
	ErrRetryStop      = errors.New("retry stop")
	ErrNotSupport     = errors.New("not support")
)

// 全双工的服务,包括服务器和客户端
type Service interface {
	Server
	Client
	Options() *Options
	Init(opts *Options) error
}

type Server interface {
	Register(srv interface{})
	Start() error
	Stop() error
	Wait() error
	Run() error
}

type Client interface {
	Send(service string, msg interface{}, opts ...CallOption) error
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

type Option func(o *Options)
type Options struct {
	// common options
	Context  context.Context
	Flags    int
	Tran     anet.Tran
	Registry registry.Registry
	Router   Router
	// server options
	Name        string         // 服务名
	Id          string         // 服务ID
	Version     string         // 服务版本
	Address     string         // Listen使用
	Advertise   string         // 注册服务使用
	BeforeStart []func() error //
	AfterStart  []func() error
	BeforeStop  []func() error
	AfterStop   []func() error
	// client options
	Selector   selector.Selector // client load balance
	Services   []string          // 需要监听的服务
	Proxy      string            // 代理服务名
	PacketFunc NewPacket         // 创建Packet
}

const (
	DefaultCallTTL = time.Second
)

type CallOption func(o *CallOptions)
type CallOptions struct {
	ID     int           // 消息ID
	Name   string        // 消息名
	Method string        // 调用函数名
	Future Future        // 异步调用
	Retry  int           // 重试次数
	TTL    time.Duration // 超时时间
}
