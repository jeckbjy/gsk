package arpc

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckbjy/micro/anet"
	_ "github.com/jeckbjy/micro/anet/tcp"
	"github.com/jeckbjy/micro/registry"
	"github.com/jeckbjy/micro/selector"
	"github.com/jeckbjy/micro/util/id"
	"github.com/jeckbjy/micro/util/options"
)

const (
	// 默认代理服务名
	ProxyService = "proxy"
)

type ServerOptions struct {
	Tran             anet.ITran
	Chain            anet.IFilterChain
	Registry         registry.IRegistry
	RegistryEnable   bool
	RegisterTTL      time.Duration
	RegisterInterval time.Duration
	Router           IRouter
	Name             string
	Id               string
	Version          string
	Address          string
	Advertise        string

	BeforeStart []func() error
	BeforeStop  []func() error
	AfterStart  []func() error
	AfterStop   []func() error
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func (o *ServerOptions) FullID() string {
	return fmt.Sprintf("%s-%s", o.Name, o.Id)
}

func (o *ServerOptions) Init(opts ...ServerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func (o *ServerOptions) SetDefaults() {
	if o.Tran == nil {
		o.Tran = anet.DefaultCreator()
		if o.Chain != nil {
			o.Tran.SetChain(o.Chain)
		} else {
			// 使用默认处理
			o.Tran.AddFilters(NewHandlerFilter())
		}
	}

	options.SetString(&o.Id, id.NewXID())
	options.SetString(&o.Name, "server")
	options.SetString(&o.Address, ":0")
	options.SetString(&o.Version, "last")
	options.SetContext(&o.Context)
}

// TODO:每个服务是否需要定制化?
type ClientOptions struct {
	Tran     anet.ITran
	Chain    anet.IFilterChain
	Registry registry.IRegistry
	Selector selector.ISelector
	Router   IRouter
	Creator  PacketCreator
	Services []string // 需要监听的服务
	Proxy    string   // 代理服务名
}

func (o *ClientOptions) Init(opts ...ClientOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func (o *ClientOptions) SetDefaults() {
}

// MiscOptions 用于Register和Send,Call
type MiscOptions struct {
	ID       int           // 消息ID
	Name     string        // 消息名字
	Method   string        // 调用函数名
	Future   IFuture       // 完成通知
	Err      error         // 失败后返回的错误
	TTL      time.Duration // 超时时间
	RetryMax int           // 最大重试次数
}

func (o *MiscOptions) Init(opts ...MiscOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// CheckFunc 重试前先校验是否需要重试,并返回新的id和时间
type CheckFunc func() (string, time.Duration, error)

// RetryFunc 尝试重新发送
type RetryFunc func() error
type RegisterRpcOptions struct {
	SeqID   string
	TTL     time.Duration
	Future  IFuture
	Retry   RetryFunc
	Check   CheckFunc
	Context context.Context // 扩展
}

type MiscOption func(*MiscOptions)
type ServerOption func(*ServerOptions)
type ClientOption func(*ClientOptions)
