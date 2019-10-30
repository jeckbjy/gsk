package base

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/util/buffer"
)

type TransferFilter struct {
	Filter
}

func (f *TransferFilter) Name() string {
	return "transfer"
}

func (f *TransferFilter) HandleWrite(ctx anet.FilterCtx) error {
	if data, ok := ctx.Data().(*buffer.Buffer); ok {
		_ = ctx.Conn().Write(data)
	}

	return nil
}
