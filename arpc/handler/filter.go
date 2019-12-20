package handler

import (
	"io"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/packet"
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/frame"
	"github.com/jeckbjy/gsk/util/buffer"
	"github.com/jeckbjy/gsk/util/errorx"
)

// 创建Filter,参数通过反射类型获得,不传则使用默认
func NewFilter(args ...interface{}) anet.Filter {
	filter := &Filter{
		frame:   frame.Default(),
		creator: packet.New,
		router:  arpc.DefaultRouter(),
		exec:    exec.Default(),
	}
	filter.pool.New = func() interface{} {
		return &Task{}
	}

	for _, a := range args {
		if a == nil {
			continue
		}
		switch v := a.(type) {
		case frame.Frame:
			filter.frame = v
		case arpc.Router:
			filter.router = v
		case arpc.NewPacket:
			filter.creator = v
		case exec.Executor:
			filter.exec = v
		case arpc.Middleware:
			filter.middleware = append(filter.middleware, v)
		case []arpc.Middleware:
			filter.middleware = append(filter.middleware, v...)
		}
	}

	return filter
}

type Filter struct {
	base.Filter
	frame      frame.Frame
	router     arpc.Router
	creator    arpc.NewPacket
	exec       exec.Executor
	pool       sync.Pool
	middleware []arpc.Middleware
}

func (f *Filter) Name() string {
	return "handler"
}

func (f *Filter) HandleRead(ctx anet.FilterCtx) error {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if !ok {
		ctx.Abort()
		return nil
	}

	// parse frame
	_, _ = buff.Seek(0, io.SeekStart)
	data, err := f.frame.Decode(buff)
	if err != nil {
		if err == frame.ErrIncomplete {
			err = nil
		}
		return err
	}

	// parse request packet
	req := f.creator()
	req.SetBuffer(data)
	if err := req.Decode(); err != nil {
		return err
	}

	handler, err := f.router.Find(req)
	if err != nil {
		return err
	}

	pcodec := codec.GetByType(req.ContentType())
	if pcodec == nil {
		return errorx.ErrNotSupport
	}

	req.SetCodec(pcodec)

	// create response packet
	rsp := f.creator()
	rsp.SetCodec(pcodec)

	// create context
	context := NewContext(ctx.Conn(), req, rsp)
	if f.exec != nil {
		task := f.pool.Get().(*Task)
		task.Init(context, handler, f.middleware)
		if err := f.exec.Handle(task); err != nil {
			return err
		}
	} else {
		if err := invoke(context, handler, f.middleware); err != nil {
			return err
		}
	}
	return nil
}

func (f *Filter) HandleWrite(ctx anet.FilterCtx) error {
	data := ctx.Data()
	var buff *buffer.Buffer
	switch v := data.(type) {
	case *buffer.Buffer:
		// 外部已经系列化好了,比如广播消息,发送效率更高
		buff = v
	case arpc.Packet:
		pcodec := codec.GetByType(v.ContentType())
		if pcodec == nil {
			return errorx.ErrNotSupport
		}

		// 消息包,序列化数据到新buffer中
		buff = buffer.New()
		v.SetCodec(pcodec)
		v.SetBuffer(buff)
		if err := v.Encode(); err != nil {
			return err
		}
		// 如果是Call,注册到router中监听回调和超时
		if info := v.CallInfo(); info != nil {
			if err := f.router.RegisterRPC(v); err != nil {
				return err
			}
		}
	default:
		return ctx.Next()
	}

	// 写入Frame
	if err := f.frame.Encode(buff); err != nil {
		return err
	}

	// 准备发送消息
	ctx.SetData(buff)
	return nil
}
