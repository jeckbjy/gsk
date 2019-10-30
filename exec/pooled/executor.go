package pooled

import (
	"math"
	"sync"

	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/exec/base"
)

func New(max int) exec.Executor {
	if max == 0 {
		max = math.MaxInt32
	}

	return &executor{max: int32(max), quit: false}
}

// 线程数不会超过max
type executor struct {
	tasks base.Queue
	mux   sync.Mutex
	wg    sync.WaitGroup
	max   int32 // 协程数不超过max
	num   int32 // 当前线程数
	quit  bool
}

func (e *executor) Stop() error {
	e.quit = true
	return nil
}

func (e *executor) Wait() {
	e.wg.Wait()
}

func (e *executor) Handle(task exec.Task) error {
	if e.quit {
		return exec.ErrAlreadyStop
	}

	create := false
	e.mux.Lock()
	e.tasks.Push(task)
	if e.num < e.max {
		e.num++
		create = true
	}
	e.mux.Unlock()

	if create {
		go e.run()
	}
	return nil
}

func (e *executor) run() {
	e.wg.Add(1)
	defer e.wg.Done()

	var task exec.Task
	for {
		e.mux.Lock()
		task = e.tasks.Pop()
		if task == nil {
			e.num--
		}
		e.mux.Unlock()
		if task == nil {
			break
		}

		_ = task.Run()
	}
}
