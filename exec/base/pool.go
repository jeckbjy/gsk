package base

import "sync"

var gPool _Pool

type _Pool struct {
	mux  sync.Mutex
	free Queue
}

func (p *_Pool) Obtain() *Node {
	p.mux.Lock()
	defer p.mux.Unlock()
	node := p.free.Pop()
	if node == nil {
		node = &Node{}
	}

	return node
}

func (p *_Pool) Free(n *Node) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.free.Push(n)
}
