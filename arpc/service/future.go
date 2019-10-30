package service

import (
	"sync"

	"github.com/jeckbjy/gsk/arpc"
)

func NewFuture() arpc.Future {
	f := &Future{}
	f.Init()
	return f
}

// Future 实现IFuture接口
type Future struct {
	count int
	cond  *sync.Cond
	mux   sync.Mutex
	err   error
}

func (f *Future) Init() {
	f.cond = sync.NewCond(&f.mux)
}

func (f *Future) Add(delta int) error {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.err != nil {
		return f.err
	}

	f.count += delta
	return nil
}

func (f *Future) Done() error {
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

func (f *Future) Fail(err error) error {
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

func (f *Future) Wait() error {
	var err error
	f.mux.Lock()
	for f.count > 0 {
		f.cond.Wait()
	}
	err = f.err
	f.mux.Unlock()
	return err
}
