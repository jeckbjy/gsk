package arpc

import (
	"context"
	"time"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/selector"
)

// 用于创建Server和Client
type Option func(o *Options)
type Options struct {
	Context   context.Context   //
	Registry  registry.Registry //
	Tran      anet.Tran         //
	Name      string            // 服务名
	Id        string            // 服务ID
	Version   string            // 服务版本
	Address   string            // Listen使用
	Advertise string            // 注册服务使用
	Selector  selector.Selector // client load balance
	Proxy     string            // 代理服务名,空代表不使用代理
}

// 重试回调函数,每次需要返回新的TTL,小于等于0则终止重试
type RetryFunc func(req Packet) time.Duration

// 用于Send,Call,Register
type MiscOption func(o *MiscOptions)
type MiscOptions struct {
	selector.Options               // 用于调用Call时,指定服务发现策略
	Method           string        // 调用方法名
	ID               int           // 消息ID,非零值
	RetryNum         int           // 重试次数
	RetryCB          RetryFunc     // 重试回调函数
	TTL              time.Duration // 超时时间
	Future           Future        // 异步等待
	Response         interface{}   // callback
	Extra            interface{}   // 自定义扩展数据
}

func (o *MiscOptions) Init(opts ...MiscOption) {
	for _, fn := range opts {
		fn(o)
	}
}

//type CallOption func(o *CallOptions)
//type CallOptions struct {
//	selector.Options
//	ID     int           // 消息ID
//	Future Future        // 异步调用
//	Retry  int           // 重试次数
//	TTL    time.Duration // 超时时间
//}
//
//func WithMsgID(msgid int) CallOption {
//	return func(o *CallOptions) {
//		o.ID = msgid
//	}
//}
//
//func WithFuture(f Future) CallOption {
//	return func(o *CallOptions) {
//		o.Future = f
//	}
//}
//
//func WithRetry(r int) CallOption {
//	return func(o *CallOptions) {
//		o.Retry = r
//	}
//}
//
//func WithTTL(ttl time.Duration) CallOption {
//	return func(o *CallOptions) {
//		o.TTL = ttl
//	}
//}
