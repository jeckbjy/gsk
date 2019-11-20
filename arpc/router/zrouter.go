package router

import (
	"github.com/jeckbjy/gsk/arpc"
)

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

func (r *Router) Register(srv interface{}, o *arpc.RegisterOptions) error {
	if o.SeqID != "" {
		return r.rpc.Register(srv, o)
	} else {
		return r.msg.Register(srv, o)
	}
}

func (r *Router) Close() error {
	r.rpc.Close()
	return nil
}
