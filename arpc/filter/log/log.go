package log

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/apm/alog"
	"github.com/jeckbjy/gsk/util/buffer"
)

const (
	Read    = 0x01
	Write   = 0x02
	Error   = 0x04
	Close   = 0x08
	Accept  = 0x10
	Connect = 0x20
	All     = 0xff
	Off     = 0x00
)

func New() anet.Filter {
	return &logFilter{mask: All}
}

type logFilter struct {
	base.Filter
	mask int
}

func (f *logFilter) Name() string {
	return "log"
}

func (f *logFilter) need(mask int) bool {
	return (f.mask & mask) != 0
}

func (f *logFilter) HandleRead(ctx anet.FilterCtx) error {
	if f.need(Read) {
		if data, ok := ctx.Data().(*buffer.Buffer); ok {
			alog.Debugf("read data:len=%+v\n", data.Len())
		}
	}

	return nil
}

func (f *logFilter) HandleWrite(ctx anet.FilterCtx) error {
	if f.need(Write) {
		if data, ok := ctx.Data().(*buffer.Buffer); ok {
			alog.Debug("send data:len=%+v", data.Len())
		}
	}
	return nil
}

func (f *logFilter) HandleOpen(ctx anet.FilterCtx) error {
	conn := ctx.Conn()
	if f.need(Connect) && conn.IsDial() {
		alog.Debugf("connect success, addr=%+v", conn.RemoteAddr())
	}

	if f.need(Accept) && !conn.IsDial() {
		alog.Debugf("accept new conn, addr=%+v", ctx.Conn().RemoteAddr())
	}

	return nil
}

func (f *logFilter) HandleClose(ctx anet.FilterCtx) error {
	if f.need(Close) {
		alog.Debugf("close conn, %+v", ctx.Conn().RemoteAddr())
	}

	return nil
}

func (f *logFilter) HandleError(ctx anet.FilterCtx) error {
	if f.need(Error) && ctx.Error() != nil {
		alog.Errorf("conn error, %+v", ctx.Error())
	}

	return nil
}
