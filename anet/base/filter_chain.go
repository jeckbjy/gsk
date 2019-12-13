package base

import "github.com/jeckbjy/gsk/anet"

func NewFilterChain() *FilterChain {
	return &FilterChain{filters: []anet.Filter{&TransferFilter{}}}
}

// 分为InBound和OutBound
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
//
type FilterChain struct {
	filters []anet.Filter
}

func (fc *FilterChain) Len() int {
	return len(fc.filters)
}

func (fc *FilterChain) Front() anet.Filter {
	if fc.Len() > 0 {
		return fc.filters[0]
	}

	return nil
}

func (fc *FilterChain) Back() anet.Filter {
	if fc.Len() > 0 {
		return fc.filters[fc.Len()-1]
	}

	return nil
}

func (fc *FilterChain) Get(index int) anet.Filter {
	return fc.filters[index]
}

func (fc *FilterChain) Index(name string) int {
	for index, filter := range fc.filters {
		if filter.Name() == name {
			return index
		}
	}

	return -1
}

func (fc *FilterChain) AddFirst(filters ...anet.Filter) {
	filters = append(filters, fc.filters[1:]...)
	fc.filters = append(fc.filters[0:0], filters...)
}

func (fc *FilterChain) AddLast(filters ...anet.Filter) {
	fc.filters = append(fc.filters, filters...)
}

func (fc *FilterChain) HandleOpen(conn anet.Conn) {
	ctx := ctxpool.New(fc.filters, conn, true, doOpen)
	if err := ctx.Call(); err != nil {
		fc.HandleError(conn, err)
	}
}

func (fc *FilterChain) HandleClose(conn anet.Conn) {
	ctx := ctxpool.New(fc.filters, conn, true, doClose)
	if err := ctx.Call(); err != nil {
		fc.HandleError(conn, err)
	}
}

func (fc *FilterChain) HandleRead(conn anet.Conn, msg interface{}) {
	ctx := ctxpool.New(fc.filters, conn, true, doRead)
	ctx.SetData(msg)
	if err := ctx.Call(); err != nil {
		fc.HandleError(conn, err)
	}
}

func (fc *FilterChain) HandleWrite(conn anet.Conn, msg interface{}) {
	ctx := ctxpool.New(fc.filters, conn, false, doWrite)
	ctx.SetData(msg)
	if err := ctx.Call(); err != nil {
		fc.HandleError(conn, err)
	}
}

func (fc *FilterChain) HandleError(conn anet.Conn, err error) {
	ctx := ctxpool.New(fc.filters, conn, false, doError)
	ctx.SetError(err)
	// 不能递归
	_ = ctx.Call()
}

func doOpen(f anet.Filter, ctx anet.FilterCtx) error {
	return f.HandleOpen(ctx)
}

func doClose(f anet.Filter, ctx anet.FilterCtx) error {
	return f.HandleClose(ctx)
}

func doRead(f anet.Filter, ctx anet.FilterCtx) error {
	return f.HandleRead(ctx)
}

func doWrite(f anet.Filter, ctx anet.FilterCtx) error {
	return f.HandleWrite(ctx)
}

func doError(f anet.Filter, ctx anet.FilterCtx) error {
	return f.HandleError(ctx)
}
