package registry

import (
	"context"
	"crypto/tls"
	"time"
)

type Option func(*Options)
type Options struct {
	Context   context.Context
	Addrs     []string
	Secure    bool
	TLSConfig *tls.Config
	Timeout   time.Duration
	TTL       time.Duration
	Interval  time.Duration
	Root      string
}

func (o *Options) Init(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

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

func WithTTL(t time.Duration) Option {
	return func(o *Options) {
		o.TTL = t
	}
}

func WithInterval(t time.Duration) Option {
	return func(o *Options) {
		o.Interval = t
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

func WithRoot(r string) Option {
	return func(o *Options) {
		o.Root = r
	}
}
