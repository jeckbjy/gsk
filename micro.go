package micro

import (
	"github.com/jeckbjy/micro/arpc"
)

func NewServer(opts ...arpc.ServerOption) arpc.IServer {
	return arpc.NewServer(opts...)
}

func NewClient(opts ...arpc.ClientOption) arpc.IClient {
	return arpc.NewClient(opts...)
}
