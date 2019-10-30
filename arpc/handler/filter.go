package handler

import (
	"io"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/frame"
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/util/buffer"
)

func NewFilter() anet.Filter {
	return &Filter{}
}

type Filter struct {
	base.Filter
	frame      frame.Frame
	codec      codec.Codec
	router     arpc.Router
	rpc        arpc.RPCRouter
	creator    arpc.NewPacket
	exec       exec.Executor
	pool       sync.Pool
	middleware []arpc.Middleware
}

func (f *Filter) Name() string {
	return "Handler"
}

func (f *Filter) HandleRead(ctx anet.FilterCtx) error {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if ok {
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

	// parse request
	req := f.creator()
	if err := req.Decode(data); err != nil {
		return err
	}

	req.SetCodec(f.codec)

	_, handler, err := f.findHandler(req)
	if err != nil {
		return err
	}

	// create response
	rsp := f.creator()
	rsp.SetCodec(f.codec)

	// create context
	context := NewContext(ctx.Conn(), req, rsp)
	task := f.pool.Get().(*Task)
	task.Init(context, handler, f.middleware)

	if f.exec != nil {
		if err := f.exec.Handle(task); err != nil {
			return err
		}
	} else {
		if err := task.Run(); err != nil {
			return err
		}
	}
	return nil
}

// 查询处理函数
func (f *Filter) findHandler(req arpc.Packet) (bool, arpc.Handler, error) {
	if req.Reply() && req.SeqID() != "" {
		h, err := f.rpc.Find(req.SeqID())
		return true, h, err
	} else {
		h, err := f.router.Find(req)
		return false, h, err
	}
}

func (f *Filter) HandleWrite(ctx anet.FilterCtx) error {
	data := ctx.Data()
	// 如果data是buffer类型,则说明外部完全托管了消息序列化
	// 对于消息广播的情况:可以外边序列化好,将Buffer保存在Body中
	// Packet:消息头每次都要单独序列化，因为每个人的消息头大概率都是不一样的
	pkg, ok := data.(arpc.Packet)
	if !ok {
		ctx.Next()
		return nil
	}

	pkg.SetCodec(f.codec)

	buff := buffer.New()

	if err := pkg.Encode(buff); err != nil {
		return err
	}

	// 写入Frame
	if err := f.frame.Encode(buff); err != nil {
		return err
	}

	// 准备发送消息
	ctx.SetData(buff)
	return nil
}
