package base

import "github.com/jeckbjy/micro/anet"

// Filter 实现空接口
type Filter struct {
}

func (f *Filter) HandleRead(ctx anet.IFilterCtx) {
}

func (f *Filter) HandleWrite(ctx anet.IFilterCtx) {
}

func (f *Filter) HandleOpen(ctx anet.IFilterCtx) {
}

func (f *Filter) HandleClose(ctx anet.IFilterCtx) {
}

func (f *Filter) HandleError(ctx anet.IFilterCtx) {
}
