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
	middleware []arpc.Middleware
	rpc        RpcRouter
	msg        MsgRouter
}

func (r *Router) Use(middleware ...arpc.Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

// 处理消息
func (r *Router) Handle(ctx arpc.Context) error {
	var handler arpc.Handler
	pkg := ctx.Message()
	if pkg.IsAck() {
		handler = r.rpc.Handle(ctx)
	} else {
		handler = r.msg.Handle(ctx)
	}

	return r.invoke(ctx, handler)
}

func (r *Router) invoke(ctx arpc.Context, handler arpc.Handler) error {
	h := handler
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}

	if h == nil {
		return arpc.ErrNoHandler
	}

	return h(ctx)
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
