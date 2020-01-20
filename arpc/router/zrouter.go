package router

import (
	"github.com/jeckbjy/gsk/arpc"
)

func New() arpc.Router {
	r := &Router{}
	r.rpc.Init()
	r.msg.Init()
	return r
}

// 默认的消息路由
type Router struct {
	handlers arpc.HandlerChain
	rpc      RpcRouter
	msg      MsgRouter
}

func (r *Router) Use(middleware ...arpc.HandlerFunc) {
	r.handlers = append(r.handlers, middleware...)
}

// 处理消息
func (r *Router) Handle(ctx arpc.Context) error {
	var handler arpc.HandlerFunc
	msg := ctx.Message()
	if msg.IsAck() && msg.SeqID() != 0 {
		handler = r.rpc.Handle(ctx)
	} else {
		handler = r.msg.Handle(ctx)
	}

	ctx.SetHandler(handler)
	ctx.SetMiddleware(r.handlers)
	return ctx.Next()
}

func (r *Router) Register(cb interface{}, opts ...arpc.MiscOption) error {
	if pkg, ok := cb.(arpc.Packet); ok {
		return r.rpc.Register(pkg)
	} else {
		o := arpc.MiscOptions{}
		o.Init(opts...)
		return r.msg.Register(cb, &o)
	}
}
