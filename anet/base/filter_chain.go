package base

import (
	"github.com/jeckbjy/micro/anet"
)

func NewFilterChain() *FilterChain {
	return &FilterChain{filters: []anet.IFilter{&TransferFilter{}}}
}

// 分为InBound和OutBound
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
type FilterChain struct {
	filters []anet.IFilter
}

func (fc *FilterChain) Len() int {
	return len(fc.filters)
}

func (fc *FilterChain) Front() anet.IFilter {
	if fc.Len() > 0 {
		return fc.filters[0]
	}

	return nil
}

func (fc *FilterChain) Back() anet.IFilter {
	if fc.Len() > 0 {
		return fc.filters[fc.Len()-1]
	}

	return nil
}

func (fc *FilterChain) Get(index int) anet.IFilter {
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

func (fc *FilterChain) AddFirst(filters ...anet.IFilter) {
	filters = append(filters, fc.filters[1:]...)
	fc.filters = append(fc.filters[0:0], filters...)
}

func (fc *FilterChain) AddLast(filters ...anet.IFilter) {
	fc.filters = append(fc.filters, filters...)
}

func (fc *FilterChain) HandleOpen(conn anet.IConn) {
	ctx := ctxpool.New(fc, conn, true, doOpen)
	ctx.Call()
}

func (fc *FilterChain) HandleClose(conn anet.IConn) {
	ctx := ctxpool.New(fc, conn, true, doClose)
	ctx.Call()
}

func (fc *FilterChain) HandleRead(conn anet.IConn, msg interface{}) {
	ctx := ctxpool.New(fc, conn, true, doRead)
	ctx.SetData(msg)
	ctx.Call()
}

func (fc *FilterChain) HandleWrite(conn anet.IConn, msg interface{}) {
	ctx := ctxpool.New(fc, conn, false, doWrite)
	ctx.SetData(msg)
	ctx.Call()
}

func (fc *FilterChain) HandleError(conn anet.IConn, err error) {
	ctx := ctxpool.New(fc, conn, false, doError)
	ctx.SetError(err)
	ctx.Call()
}

func doOpen(f anet.IFilter, ctx anet.IFilterCtx) {
	f.HandleOpen(ctx)
}

func doClose(f anet.IFilter, ctx anet.IFilterCtx) {
	f.HandleClose(ctx)
}

func doRead(f anet.IFilter, ctx anet.IFilterCtx) {
	f.HandleRead(ctx)
}

func doWrite(f anet.IFilter, ctx anet.IFilterCtx) {
	f.HandleWrite(ctx)
}

func doError(f anet.IFilter, ctx anet.IFilterCtx) {
	f.HandleError(ctx)
}
