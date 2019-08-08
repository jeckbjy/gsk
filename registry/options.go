package registry

import (
	"context"
	"crypto/tls"
	"time"
)

type Options struct {
	Addrs     []string
	Timeout   time.Duration
	Secure    bool
	TLSConfig *tls.Config
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func (o *Options) Init(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

type RegisterOptions struct {
	TTL      time.Duration // 过期时间,不能为0
	Interval time.Duration // 自动注册时间间隔,默认TTL的一半
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func (o *RegisterOptions) Init(opts ...RegisterOption) {
	o.TTL = time.Second
	for _, opt := range opts {
		opt(o)
	}

	if o.Interval == 0 {
		o.Interval = o.TTL / 2
	}
}

type WatchOptions struct {
	// Specify a service to watch
	// If blank, the watch is for all services
	Services []string
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func (o *WatchOptions) Init(opts ...WatchOption) {
	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*Options)
type WatchOption func(*WatchOptions)
type RegisterOption func(*RegisterOptions)

// WithAddrs is the registry addresses to use
func WithAddrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

func WithTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// WithSecure communication with the registry
func WithSecure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// Specify TLS Config
func WithTLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

func WithRegisterTTL(t time.Duration) RegisterOption {
	return func(o *RegisterOptions) {
		if t > 0 {
			o.TTL = t
		}
	}
}

func WithRegisterInterval(t time.Duration) RegisterOption {
	return func(options *RegisterOptions) {
		if t > 0 {
			options.Interval = t
		}
	}
}

// Watch services
func WithWatchServices(services ...string) WatchOption {
	return func(o *WatchOptions) {
		o.Services = services
	}
}
