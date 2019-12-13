package base

import (
	"errors"
	"fmt"
	"math"

	"github.com/jeckbjy/gsk/anet"
)

var ErrIndexOverflow = errors.New("filter index overflow")

// callback 用于Next等调用中，能执行对应的HandleRead，HandleWrite函数
type callback func(filter anet.Filter, ctx anet.FilterCtx) error

// FilterCtx implement anet.FilterCtx
type FilterCtx struct {
	next *FilterCtx // 用于pool中存储空闲链表
	//chain   anet.FilterChain
	filters []anet.Filter
	conn    anet.Conn
	data    interface{}
	err     error
	index   int
	forward bool
	cb      callback
}

func (ctx *FilterCtx) init(filters []anet.Filter, conn anet.Conn, forward bool, cb callback) {
	//ctx.chain = chain
	ctx.filters = filters
	ctx.conn = conn
	ctx.data = nil
	ctx.forward = forward
	ctx.cb = cb
	// 调用next时会自动加一或减一
	if forward {
		ctx.index = -1
	} else {
		ctx.index = len(filters)
	}
}

func (ctx *FilterCtx) Conn() anet.Conn {
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

func (ctx *FilterCtx) Abort() {
	if ctx.forward {
		ctx.index = math.MaxInt32
	} else {
		ctx.index = -1
	}
}

// 调用callback
func (ctx *FilterCtx) call(index int) error {
	ctx.index = index
	// 这样写可以保证即使没有调用Next也能正确执行到最后一个Filter
	// 如果需要终止，需要主动调用Abort
	//
	// Open,Close,Read是forward
	if ctx.forward {
		for idx := len(ctx.filters); ctx.index < idx; ctx.index++ {
			if err := ctx.cb(ctx.filters[ctx.index], ctx); err != nil {
				ctx.Abort()
				return err
			}
		}
	} else {
		for ; ctx.index >= 0; ctx.index-- {
			if err := ctx.cb(ctx.filters[ctx.index], ctx); err != nil {
				ctx.Abort()
				return err
			}
		}
	}

	return nil
}

func (ctx *FilterCtx) Next() error {
	if ctx.forward {
		return ctx.call(ctx.index + 1)
	} else {
		return ctx.call(ctx.index - 1)
	}
}

func (ctx *FilterCtx) Jump(index int) error {
	if index < 0 || index >= len(ctx.filters) {
		return ErrIndexOverflow
	}

	return ctx.call(index)
}

func (ctx *FilterCtx) JumpBy(name string) error {
	index := -1
	for idx, filter := range ctx.filters {
		if filter.Name() == name {
			index = idx
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("not found filter,%+v", name)
	}

	return ctx.Jump(index)
}

// 克隆一份,可以用于转移到其他线程继续执行
func (ctx *FilterCtx) Clone() anet.FilterCtx {
	nctx := ctxpool.New(ctx.filters, ctx.conn, ctx.forward, ctx.cb)
	nctx.data = ctx.data
	nctx.err = ctx.err
	nctx.index = ctx.index
	return nctx
}

func (ctx *FilterCtx) Call() error {
	err := ctx.Next()
	ctxpool.Free(ctx)
	return err
}
