package main

import (
	"encoding/json"
	"io"
	"log"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/util/buffer"
)

type HandlerFilter struct {
	base.Filter
	Peer string
}

func (f *HandlerFilter) Name() string {
	return "HandlerFilter"
}

func (f *HandlerFilter) HandleRead(ctx anet.IFilterCtx) {
	data := ctx.Data().(*buffer.Buffer)
	// decode
	data.Seek(0, io.SeekStart)
	decoder := json.NewDecoder(data)
	req := &EchoMsg{}
	if err := decoder.Decode(req); err != nil {
		log.Printf("decode fail:%+v,%s, %+v", data.Len(), data.String(), err)
		return
	}

	// process message
	rsp := &EchoMsg{}
	if req.Text == "ping" {
		rsp.Text = "pong"
	} else {
		rsp.Text = "ping"
	}

	log.Printf("recv: %s", req.Text)
	log.Printf("send: %s", rsp.Text)
	_ = ctx.Conn().Send(rsp)
}

func (f *HandlerFilter) HandleWrite(ctx anet.IFilterCtx) {
	buff := buffer.New()
	encoder := json.NewEncoder(buff)
	if err := encoder.Encode(ctx.Data()); err != nil {
		ctx.Abort(err)
		return
	}

	ctx.SetData(buff)
}
