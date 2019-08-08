package arpc

import (
	"github.com/jeckbjy/micro/anet"
	"github.com/jeckbjy/micro/registry"
)

func WithSTran(tran anet.ITran) ServerOption {
	return func(opts *ServerOptions) {
		opts.Tran = tran
	}
}

func WithSChain(chain anet.IFilterChain) ServerOption {
	return func(opts *ServerOptions) {
		opts.Chain = chain
	}
}

func WithSRegistry(r registry.IRegistry) ServerOption {
	return func(opts *ServerOptions) {
		opts.Registry = r
	}
}

func WithSRouter(r IRouter) ServerOption {
	return func(opts *ServerOptions) {
		opts.Router = r
	}
}

func WithName(name string) ServerOption {
	return func(opts *ServerOptions) {
		opts.Name = name
	}
}

func WithID(id string) ServerOption {
	return func(opts *ServerOptions) {
		opts.Id = id
	}
}

func WithVersion(version string) ServerOption {
	return func(opts *ServerOptions) {
		opts.Version = version
	}
}

func WithAddress(address string) ServerOption {
	return func(opts *ServerOptions) {
		opts.Address = address
	}
}
