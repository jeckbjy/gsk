package gsk

import (
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/service"
)

func New(opts ...arpc.Option) (arpc.Service, error) {
	o := &arpc.Options{}
	for _, fn := range opts {
		fn(o)
	}
	return service.New(o)
}

func Name(name string) arpc.Option {
	return func(o *arpc.Options) {
		o.Name = name
	}
}

func ID(id string) arpc.Option {
	return func(o *arpc.Options) {
		o.Id = id
	}
}
