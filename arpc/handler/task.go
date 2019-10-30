package handler

import "github.com/jeckbjy/gsk/arpc"

type Task struct {
	ctx        arpc.IContext
	handler    arpc.Handler
	middleware []arpc.Middleware
}

func (t *Task) Init(ctx arpc.IContext, handler arpc.Handler, middleware []arpc.Middleware) {
	t.ctx = ctx
	t.handler = handler
	t.middleware = middleware
}

func (t *Task) Run() error {
	h := t.handler
	for i := len(t.middleware); i >= 0; i-- {
		h = t.middleware[i](h)
	}

	return h(t.ctx)
}
