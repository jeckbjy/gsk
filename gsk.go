package gsk

import (
	"github.com/jeckbjy/gsk/arpc"
)

func NewServer(opts ...arpc.ServerOption) arpc.IServer {
	return arpc.NewServer(opts...)
}

func NewClient(opts ...arpc.ClientOption) arpc.IClient {
	return arpc.NewClient(opts...)
}
