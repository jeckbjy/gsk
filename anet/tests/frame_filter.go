package main

import (
	"encoding/binary"
	"io"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/util/buffer"
)

type FrameFilter struct {
	base.Filter
}

func (f *FrameFilter) Name() string {
	return "FrameFilter"
}

func (f *FrameFilter) HandleRead(ctx anet.FilterCtx) error {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if !ok {
		ctx.Abort()
		return nil
	}

	_, _ = buff.Seek(0, io.SeekStart)
	// frame 粘包处理,两个字节
	var data [2]byte
	if n, _ := buff.Read(data[:]); n != 2 {
		ctx.Abort()
		return nil
	}

	var length = binary.LittleEndian.Uint16(data[:])
	if buff.Len()-buff.Pos() < int(length) {
		ctx.Abort()
		return nil
	}

	// 去除长度信息
	buff.Discard()
	buff.Seek(int64(length), io.SeekStart)

	msgdata := buff.Split()
	ctx.SetData(msgdata)
	return nil
}

func (f *FrameFilter) HandleWrite(ctx anet.FilterCtx) error {
	buf := ctx.Data().(*buffer.Buffer)
	len := buf.Len()
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(len))
	buf.Prepend(data)
	return nil
}
