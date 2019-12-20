package exec

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
)

func New() anet.Filter {
	f := &execFilter{}
	return f
}

type execFilter struct {
	base.Filter
}

func (f *execFilter) Name() string {
	return "exec"
}

func (f *execFilter) HandleRead(ctx anet.FilterCtx) error {
	return nil
}
