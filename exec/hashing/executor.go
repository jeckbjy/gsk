package hashing

import (
	"sync"

	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/exec/base"
)

func New(strategy Strategy, maxWorker int) exec.Executor {
	e := &executor{strategy: strategy}
	if maxWorker > 0 {
		e.workers = make([]*base.Worker, maxWorker)
	}
	return e
}

type Strategy func(task exec.Task) int

// 根据hash的方式投递到某个线程中执行
type executor struct {
	workers  []*base.Worker
	strategy Strategy
	mux      sync.Mutex
	quit     bool
	wg       sync.WaitGroup
}

func (e *executor) Stop() error {
	e.mux.Lock()
	e.quit = true
	for _, w := range e.workers {
		w.Stop()
	}
	e.mux.Unlock()

	return nil
}

func (e *executor) Wait() {
	e.wg.Wait()
}

func (e *executor) Post(task exec.Task) error {
	if e.quit {
		return exec.ErrAlreadyStop
	}

	index := e.strategy(task)
	worker := e.obtain(index)
	worker.Post(task)

	return nil
}

func (e *executor) obtain(index int) *base.Worker {
	var worker *base.Worker

	e.mux.Lock()
	if index >= len(e.workers) {
		n := make([]*base.Worker, index+1)
		copy(n, e.workers)
	}

	if e.workers[index] == nil {
		worker := base.NewWorker()
		worker.Start(&e.wg)
		e.workers[index] = worker
	}

	worker = e.workers[index]

	e.mux.Unlock()
	return worker
}
