package rc4

import (
	"crypto/rc4"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/util/buffer"
)

func New(key []byte) anet.Filter {
	return &rc4Filter{}
}

type rc4Filter struct {
	base.Filter
	key []byte
}

func (*rc4Filter) Name() string {
	return "rc4"
}

func (f *rc4Filter) HandleRead(ctx anet.FilterCtx) error {
	return f.process(ctx)
}

func (f *rc4Filter) HandleWrite(ctx anet.FilterCtx) error {
	return f.process(ctx)
}

func (f *rc4Filter) process(ctx anet.FilterCtx) error {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if ok {
		// rc4是有状态的,每次都需要新建
		cipher, err := rc4.NewCipher(f.key)
		if err != nil {
			return err
		}

		buff.Visit(func(data []byte) bool {
			cipher.XORKeyStream(data, data)
			return true
		})
	}

	return nil
}
