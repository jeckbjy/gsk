package server

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/registry"
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

func Name(n string) arpc.Option {
	return func(o *arpc.Options) {
		o.Name = n
	}
}

func ID(id string) arpc.Option {
	return func(o *arpc.Options) {
		o.Id = id
	}
}

func Address(addr string) arpc.Option {
	return func(o *arpc.Options) {
		o.Address = addr
	}
}

func Advertise(addr string) arpc.Option {
	return func(o *arpc.Options) {
		o.Advertise = addr
	}
}
