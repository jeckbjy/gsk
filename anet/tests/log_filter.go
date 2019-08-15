package main

import (
	"log"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/util/buffer"
)

type LogFilter struct {
	base.Filter
}

func (f *LogFilter) Name() string {
	return "LogFilter"
}

func (f *LogFilter) HandleRead(ctx anet.IFilterCtx) {
	if buff, ok := ctx.Data().(*buffer.Buffer); ok {
		log.Printf("recv data:%+v", buff.Len())
	}
}

func (f *LogFilter) HandleWrite(ctx anet.IFilterCtx) {
	if buff, ok := ctx.Data().(*buffer.Buffer); ok {
		log.Printf("send data:%+v", buff.Len())
	}
}
