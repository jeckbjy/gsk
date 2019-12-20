package frame

import (
	"io"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/frame"
	"github.com/jeckbjy/gsk/util/buffer"
)

type Option func(f *frameFilter)

func Frame(frame frame.Frame) Option {
	return func(f *frameFilter) {
		f.frame = frame
	}
}

func New(opts ...Option) anet.Filter {
	f := &frameFilter{frame: frame.Default()}
	for _, fn := range opts {
		fn(f)
	}

	return f
}

// 用于粘包处理
type frameFilter struct {
	base.Filter
	frame frame.Frame
}

func (f *frameFilter) Name() string {
	return "frame"
}

func (f *frameFilter) HandleRead(ctx anet.FilterCtx) error {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if !ok {
		ctx.Abort()
		return nil
	}

	_, _ = buff.Seek(0, io.SeekStart)
	data, err := f.frame.Decode(buff)
	if err != nil {
		if err == frame.ErrIncomplete {
			err = nil
		}
		return err
	}

	ctx.SetData(data)
	return nil
}

func (f *frameFilter) HandleWrite(ctx anet.FilterCtx) error {
	if data, ok := ctx.Data().(*buffer.Buffer); ok {
		return f.frame.Encode(data)
	}

	return nil
}
