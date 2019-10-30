package gsk

import (
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/service"
)

func New(opts ...arpc.Option) (arpc.Service, error) {
	return service.New(opts...)
}
