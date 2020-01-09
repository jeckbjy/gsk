package fexec

import (
	"sync"

	"github.com/jeckbjy/gsk/arpc"
)

var gTaskPool = sync.Pool{
	New: func() interface{} {
		return &Task{}
	},
}

func newTask(ctx arpc.Context, router arpc.Router) *Task {
	task := gTaskPool.Get().(*Task)
	task.Init(ctx, router)
	return task
}

type Task struct {
	ctx    arpc.Context
	router arpc.Router
}

func (t *Task) Init(ctx arpc.Context, router arpc.Router) {
	t.ctx = ctx
	t.router = router
}

func (t *Task) Run() error {
	err := t.router.Handle(t.ctx)
	t.ctx.Free()
	gTaskPool.Put(t)
	return err
}
