package arpc

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc/selector"
	"github.com/jeckbjy/gsk/registry"
)

func WithCTran(tran anet.ITran) ClientOption {
	return func(opts *ClientOptions) {
		opts.Tran = tran
	}
}

func WithCChain(chain anet.IFilterChain) ClientOption {
	return func(opts *ClientOptions) {
		opts.Chain = chain
	}
}

func WithCRegistry(r registry.IRegistry) ClientOption {
	return func(opts *ClientOptions) {
		opts.Registry = r
	}
}

func WithCRouter(r IRouter) ClientOption {
	return func(opts *ClientOptions) {
		opts.Router = r
	}
}

func WithCSelector(s selector.ISelector) ClientOption {
	return func(opts *ClientOptions) {
		opts.Selector = s
	}
}

func WithCCreator(c PacketCreator) ClientOption {
	return func(opts *ClientOptions) {
		opts.Creator = c
	}
}

func WithCServices(services ...string) ClientOption {
	return func(opts *ClientOptions) {
		opts.Services = services
	}
}

func WithCProxy(proxy string) ClientOption {
	if proxy == "" {
		proxy = "proxy"
	}
	return func(opts *ClientOptions) {
		opts.Proxy = proxy
	}
}
