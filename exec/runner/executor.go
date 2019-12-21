package runner

import "github.com/jeckbjy/gsk/exec"

func New() exec.Executor {
	e := &executor{}
	return e
}

// 直接运行task
type executor struct {
}

func (e *executor) Post(task exec.Task) error {
	return task.Run()
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) Wait() {
}
