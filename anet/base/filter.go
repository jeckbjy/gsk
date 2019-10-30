package base

import "github.com/jeckbjy/gsk/anet"

// Filter 实现空接口
type Filter struct {
}

func (f *Filter) HandleRead(ctx anet.FilterCtx) error {
	return nil
}

func (f *Filter) HandleWrite(ctx anet.FilterCtx) error {
	return nil
}

func (f *Filter) HandleOpen(ctx anet.FilterCtx) error {
	return nil
}

func (f *Filter) HandleClose(ctx anet.FilterCtx) error {
	return nil
}

func (f *Filter) HandleError(ctx anet.FilterCtx) error {
	return nil
}
