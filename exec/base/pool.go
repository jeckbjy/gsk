package base

import "sync"

var gPool _Pool

type _Pool struct {
	mux  sync.Mutex
	free Queue
}

func (p *_Pool) Obtain() *Node {
	p.mux.Lock()
	node := p.free.popNode()
	if node == nil {
		node = &Node{}
	}
	p.mux.Unlock()

	return node
}

func (p *_Pool) Free(n *Node) {
	p.mux.Lock()
	p.free.pushNode(n)
	p.mux.Unlock()
}
