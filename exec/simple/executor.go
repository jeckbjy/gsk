package simple

import (
	"sync"

	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/exec/base"
)

func New() exec.Executor {
	e := &executor{}
	e.Start()
	return e
}

// 将task投递到另外一个单独线程中执行
type executor struct {
	worker *base.Worker
	wg     sync.WaitGroup
}

func (e *executor) Start() {
	e.worker = base.NewWorker()
	e.worker.Start(&e.wg)
}

func (e *executor) Stop() error {
	e.worker.Stop()
	return nil
}

func (e *executor) Wait() {
	e.wg.Wait()
}

func (e *executor) Post(task exec.Task) error {
	e.worker.Post(task)
	return nil
}
