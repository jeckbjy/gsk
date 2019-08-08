package base

import (
	"errors"
	"github.com/jeckbjy/micro/anet"
	"math"
)

var ErrIndexOverflow = errors.New("filter index overflow")

// callback 用于Next等调用中，能执行对应的HandleRead，HandleWrite函数
type callback func(filter anet.IFilter, ctx anet.IFilterCtx)

// FilterCtx implement anet.FilterCtx
type FilterCtx struct {
	next    *FilterCtx // 用于pool中存储空闲链表
	chain   anet.IFilterChain
	conn    anet.IConn
	data    interface{}
	err     error
	index   int
	forward bool
	cb      callback
}

func (ctx *FilterCtx) init(chain anet.IFilterChain, conn anet.IConn, forward bool, cb callback) {
	ctx.chain = chain
	ctx.conn = conn
	ctx.data = nil
	ctx.forward = forward
	ctx.cb = cb
	if forward {
		ctx.index = -1
	} else {
		ctx.index = chain.Len()
	}
}

func (ctx *FilterCtx) Conn() anet.IConn {
	return ctx.conn
}

func (ctx *FilterCtx) Data() interface{} {
	return ctx.data
}

func (ctx *FilterCtx) SetData(data interface{}) {
	ctx.data = data
}

func (ctx *FilterCtx) Error() error {
	return ctx.err
}

func (ctx *FilterCtx) SetError(err error) {
	ctx.err = err
}

func (ctx *FilterCtx) IsAbort() bool {
	return ctx.index >= math.MaxInt32
}

func (ctx *FilterCtx) Abort(err error) {
	ctx.index = math.MaxInt32
	if err != nil {
		ctx.chain.HandleError(ctx.conn, err)
	}
}

// 调用callback
func (ctx *FilterCtx) call(index int) {
	ctx.index = index
	// 这样写可以保证即使没有调用Next也能正确执行到最后一个Filter
	// 如果需要终止，需要主动调用Abort
	if ctx.forward {
		for idx := ctx.chain.Len(); ctx.index < idx; ctx.index++ {
			ctx.cb(ctx.chain.Get(ctx.index), ctx)
		}
	} else {
		for ; ctx.index >= 0; ctx.index-- {
			ctx.cb(ctx.chain.Get(ctx.index), ctx)
		}
	}
}

func (ctx *FilterCtx) Next() {
	if ctx.forward {
		ctx.call(ctx.index + 1)
	} else {
		ctx.call(ctx.index - 1)
	}
}

func (ctx *FilterCtx) Jump(index int) error {
	if index < 0 || index >= ctx.chain.Len() {
		return ErrIndexOverflow
	}

	ctx.call(index)
	return nil
}

func (ctx *FilterCtx) JumpBy(name string) error {
	index := ctx.chain.Index(name)
	return ctx.Jump(index)
}

func (ctx *FilterCtx) Clone() anet.IFilterCtx {
	nctx := ctxpool.New(ctx.chain, ctx.conn, ctx.forward, ctx.cb)
	nctx.data = ctx.data
	nctx.err = ctx.err
	nctx.index = ctx.index
	return nctx
}

func (ctx *FilterCtx) Call() {
	ctx.Next()
	ctxpool.Free(ctx)
}
