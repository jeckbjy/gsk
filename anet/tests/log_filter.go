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

func (f *LogFilter) HandleRead(ctx anet.FilterCtx) error {
	if buff, ok := ctx.Data().(*buffer.Buffer); ok {
		log.Printf("log read, recv data:%+v", buff.Len())
	}

	return nil
}

func (f *LogFilter) HandleWrite(ctx anet.FilterCtx) error {
	if buff, ok := ctx.Data().(*buffer.Buffer); ok {
		log.Printf("log write,send data:%+v", buff.Len())
	}

	return nil
}

func (f *LogFilter) HandleError(ctx anet.FilterCtx) error {
	log.Printf("some err,%+v", ctx.Error())

	return nil
}
