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
	Send(service string, req interface{}, opts ...CallOption) error
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
