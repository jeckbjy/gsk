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

func (f *_Future) Add() {
	f.mux.Lock()
	f.count++
	f.mux.Unlock()
}

func (f *_Future) Done(err error) {
	f.mux.Lock()
	f.count--

	if f.err == nil && err != nil {
		f.err = err
	}

	notify := f.count <= 0 || f.err != nil
	f.mux.Unlock()

	if notify {
		f.cond.Signal()
	}
}

func (f *_Future) Wait() error {
	f.mux.Lock()
	for f.count > 0 && f.err == nil {
		f.cond.Wait()
	}
	err := f.err
	f.mux.Unlock()
	return err
}
