package logging

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
			alog.Debugf("read data:len=%+v,connID=%+v\n", data.Len(), ctx.Conn().ID())
		}
	}

	return nil
}

func (f *logFilter) HandleWrite(ctx anet.FilterCtx) error {
	if f.need(Write) {
		if data, ok := ctx.Data().(*buffer.Buffer); ok {
			alog.Debugf("send data:len=%+v,connID=%+v", data.Len(), ctx.Conn().ID())
		}
	}
	return nil
}

func (f *logFilter) HandleOpen(ctx anet.FilterCtx) error {
	conn := ctx.Conn()
	if f.need(Connect) && conn.IsDial() {
		alog.Debugf("connect success, addr=%+v,connID=%+v", conn.RemoteAddr(), ctx.Conn().ID())
	}

	if f.need(Accept) && !conn.IsDial() {
		alog.Debugf("accept new conn, addr=%+v,connID=%+v", ctx.Conn().RemoteAddr(), ctx.Conn().ID())
	}

	return nil
}

func (f *logFilter) HandleClose(ctx anet.FilterCtx) error {
	if f.need(Close) {
		alog.Debugf("close conn, addr=%+v,connID=%+v", ctx.Conn().RemoteAddr(), ctx.Conn().ID())
	}

	return nil
}

func (f *logFilter) HandleError(ctx anet.FilterCtx) error {
	if f.need(Error) && ctx.Error() != nil {
		alog.Errorf("conn error, err=%+v,connID=%+v", ctx.Error(), ctx.Conn().ID())
	}

	return nil
}
