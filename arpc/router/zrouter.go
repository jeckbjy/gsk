package router

import (
	"github.com/jeckbjy/gsk/arpc"
)

func init() {
	arpc.SetDefaultRouter(New())
}

func New() arpc.Router {
	r := &Router{}
	r.msg.Init()
	r.rpc.Init()
	return r
}

type Router struct {
	msg _MsgRouter
	rpc _RpcRouter
}

// 查询消息回调
func (r *Router) Find(pkg arpc.Packet) (arpc.Handler, error) {
	if pkg.IsAck() && pkg.SeqID() != "" {
		return r.rpc.Find(pkg)
	} else {
		return r.msg.Find(pkg)
	}
}

func (r *Router) Register(srv interface{}, opts ...arpc.RegisterOption) error {
	o := arpc.RegisterOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	return r.msg.Register(srv, &o)
}

func (r *Router) RegisterRPC(req arpc.Packet) error {
	return r.rpc.Register(req)
}

func (r *Router) Close() error {
	r.rpc.Close()
	return nil
}
