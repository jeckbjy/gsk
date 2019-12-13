package gsk

import (
	"context"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/client"
	"github.com/jeckbjy/gsk/arpc/handler"
	"github.com/jeckbjy/gsk/arpc/server"
	"github.com/jeckbjy/gsk/broker"
	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/selector"
	rselector "github.com/jeckbjy/gsk/selector/registry"
)

func newOptions(name string, opts ...Option) *Options {
	o := &Options{Context: context.Background(), Name: name, Router: arpc.DefaultRouter()}
	for _, fn := range opts {
		fn(o)
	}

	if o.Server != nil && o.Client != nil {
		return o
	}

	// 使用默认值创建
	regi := registry.Default()
	tran := anet.Default()()
	if len(o.Filters) == 0 {
		tran.AddFilters(handler.NewFilter(o.Middleware, o.Exec, o.Router))
	} else {
		tran.AddFilters(o.Filters...)
	}
	if o.Selector == nil {
		o.Selector = rselector.New(regi)
	}

	if o.Server == nil {
		o.Server = server.New(
			server.Name(name),
			server.Registry(regi),
			server.Transport(tran),
			server.ID(o.Id),
			server.Address(o.Address),
			server.Advertise(o.Advertise),
		)
	}

	if o.Client == nil {
		o.Client = client.New(
			client.Registry(regi),
			client.Transport(tran),
			client.Proxy(o.Proxy),
			client.Selector(o.Selector),
		)
	}

	return o
}

type Callback func() error
type Option func(o *Options)

// 系统默认提供了一些默认设置,如果需要扩展,自定义,可通过以下方式
// 一:手动设置Server和Client,这种方式最繁琐,所有的配置项都需要自己设置,注意:如果使用这种方式,系统将会忽略默认配置,完全由用户托管
// 二:
//	1:每个使用到的组件都提供了SetDefault选项用于全局替换组件,只需要初始化的地方设置一下即可,比如registry,selector,anet等
// 	2:有一些会经常用到的参数可直接在Options中设置,用于初始化Server和Client,如Proxy
type Options struct {
	Context     context.Context
	BeforeStart []Callback
	AfterStart  []Callback
	BeforeStop  []Callback
	AfterStop   []Callback
	Broker      broker.Broker
	Server      arpc.Server
	Client      arpc.Client
	Router      arpc.Router
	Selector    selector.Selector
	Exec        exec.Executor
	Filters     []anet.Filter
	Middleware  []arpc.Middleware
	Name        string
	Id          string
	Address     string
	Advertise   string
	Proxy       string
}

func Context(c context.Context) Option {
	return func(o *Options) {
		o.Context = c
	}
}

func BeforeStart(cb Callback) Option {
	return func(o *Options) {
		o.BeforeStart = append(o.BeforeStart, cb)
	}
}

func AfterStart(cb Callback) Option {
	return func(o *Options) {
		o.AfterStart = append(o.AfterStart, cb)
	}
}

func BeforeStop(cb Callback) Option {
	return func(o *Options) {
		o.BeforeStop = append(o.BeforeStop, cb)
	}
}

func AfterStop(cb Callback) Option {
	return func(o *Options) {
		o.AfterStop = append(o.AfterStop, cb)
	}
}

func Server(s arpc.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}

func Client(c arpc.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

func Executor(e exec.Executor) Option {
	return func(o *Options) {
		o.Exec = e
	}
}

func Filter(f ...anet.Filter) Option {
	return func(o *Options) {
		o.Filters = f
	}
}

func Middleware(m ...arpc.Middleware) Option {
	return func(o *Options) {
		o.Middleware = m
	}
}

func ServiceID(id string) Option {
	return func(o *Options) {
		o.Id = id
	}
}

func Address(addr string) Option {
	return func(o *Options) {
		o.Address = addr
	}
}

func Advertise(addr string) Option {
	return func(o *Options) {
		o.Advertise = addr
	}
}

func Proxy(proxy string) Option {
	return func(o *Options) {
		o.Proxy = proxy
	}
}
