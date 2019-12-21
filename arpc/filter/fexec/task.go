package fexec

import (
	"sync"

	"github.com/jeckbjy/gsk/arpc"
)

type Task struct {
	pool       *sync.Pool
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
	err := invoke(t.ctx, t.handler, t.middleware)
	t.pool.Put(t)
	t.ctx.Free()
	return err
}

func invoke(ctx arpc.Context, handler arpc.Handler, middleware []arpc.Middleware) error {
	h := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	return h(ctx)
}
