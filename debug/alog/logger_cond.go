package alog

import "sync"

type Cond struct {
	cond *sync.Cond
	mux  sync.Mutex
}

func (c *Cond) Init() {
	c.cond = sync.NewCond(&c.mux)
}

func (c *Cond) Lock() {
	c.mux.Lock()
}

func (c *Cond) Unlock() {
	c.mux.Unlock()
}

func (c *Cond) Wait() {
	c.cond.Wait()
}

func (c *Cond) Signal() {
	c.cond.Signal()
}
