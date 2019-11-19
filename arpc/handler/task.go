package handler

import "github.com/jeckbjy/gsk/arpc"

func invoke(ctx arpc.Context, handler arpc.Handler, middleware []arpc.Middleware) error {
	h := handler
	for i := len(middleware); i >= 0; i-- {
		h = middleware[i](h)
	}

	return h(ctx)
}

type Task struct {
	ctx        arpc.Context
	handler    arpc.Handler
	middleware []arpc.Middleware
}

func (t *Task) Init(ctx arpc.Context, handler arpc.Handler, middleware []arpc.Middleware) {
	t.ctx = ctx
	t.handler = handler
	t.middleware = middleware
}

func (t *Task) Run() error {
	return invoke(t.ctx, t.handler, t.middleware)
}
