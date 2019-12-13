package client

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/selector"
)

func Registry(r registry.Registry) arpc.Option {
	return func(o *arpc.Options) {
		o.Registry = r
	}
}

func Transport(t anet.Tran) arpc.Option {
	return func(o *arpc.Options) {
		o.Tran = t
	}
}

func Selector(s selector.Selector) arpc.Option {
	return func(o *arpc.Options) {
		o.Selector = s
	}
}

func Proxy(p string) arpc.Option {
	return func(o *arpc.Options) {
		o.Proxy = p
	}
}
