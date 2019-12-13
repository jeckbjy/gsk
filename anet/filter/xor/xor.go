package xor

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/util/buffer"
)

func New(key []byte) anet.Filter {
	return &xorFilter{key: key}
}

// 基于xor方法的加密
// TODO:增加标识忽略某些协议加密,比如Auth,Login,HealthCheck等
type xorFilter struct {
	base.Filter
	key []byte
}

func (*xorFilter) Name() string {
	return "xor"
}

func (f *xorFilter) HandleRead(ctx anet.FilterCtx) error {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if ok {
		f.process(buff)
	}

	return nil
}

func (f *xorFilter) HandleWrite(ctx anet.FilterCtx) error {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if ok {
		f.process(buff)
	}

	return nil
}

func (f *xorFilter) process(buff *buffer.Buffer) {
	length := len(f.key)
	index := 0
	buff.Visit(func(data []byte) bool {
		for i, v := range data {
			index++
			data[i] = v ^ f.key[index%length]
		}
		return true
	})
}
