package client

import (
	"sync"

	"github.com/jeckbjy/gsk/arpc"
)

func NewFuture() arpc.Future {
	f := &_Future{}
	f.Init()
	return f
}

// _Future 实现Future接口
type _Future struct {
	count int
	cond  *sync.Cond
	mux   sync.Mutex
	err   error
}

func (f *_Future) Init() {
	f.cond = sync.NewCond(&f.mux)
}

func (f *_Future) Add(delta int) error {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.err != nil {
		return f.err
	}

	f.count += delta
	return nil
}

func (f *_Future) Done() error {
	f.mux.Lock()
	if f.err != nil {
		f.mux.Unlock()
		return f.err
	}

	f.count--
	f.mux.Unlock()
	if f.count == 0 {
		f.cond.Signal()
	}
	return nil
}

func (f *_Future) Fail(err error) error {
	f.mux.Lock()
	if f.err != nil {
		f.mux.Unlock()
		return f.err
	}
	f.err = err
	f.count = 0
	f.mux.Unlock()
	f.cond.Signal()
	return nil
}

func (f *_Future) Wait() error {
	var err error
	f.mux.Lock()
	for f.count > 0 {
		f.cond.Wait()
	}
	err = f.err
	f.mux.Unlock()
	return err
}
