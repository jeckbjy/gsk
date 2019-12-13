package base

import (
	"sync"

	"github.com/jeckbjy/gsk/anet"
)

var ctxpool = &FilterCtxPool{}

// FilterCtxPool Context缓存,TODO:自动收缩
type FilterCtxPool struct {
	frees *FilterCtx
	mux   sync.Mutex
	count int
}

func (p *FilterCtxPool) New(filters []anet.Filter, conn anet.Conn, forward bool, cb callback) *FilterCtx {
	p.mux.Lock()
	defer p.mux.Unlock()
	var ctx *FilterCtx
	if p.frees != nil {
		ctx = p.frees
		p.frees = ctx.next
		ctx.next = nil
		p.count--
	} else {
		ctx = &FilterCtx{}
	}

	ctx.init(filters, conn, forward, cb)
	return ctx
}

func (p *FilterCtxPool) Free(ctx *FilterCtx) {
	p.mux.Lock()
	defer p.mux.Unlock()
	ctx.next = p.frees
	p.frees = ctx
	p.count++
}
