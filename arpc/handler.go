package arpc

import (
	"io"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/arpc/codec"
	"github.com/jeckbjy/gsk/arpc/frame"
	"github.com/jeckbjy/gsk/util/buffer"
)

func NewHandlerFilter(opts ...HandlerFilterOption) anet.IFilter {
	f := &HandlerFilter{}
	for _, opt := range opts {
		opt(f)
	}
	if f.frame == nil {
		f.frame = frame.Default
	}
	if f.codec == nil {
		f.codec = codec.Default
	}
	if f.creator == nil {
		f.creator = NewPacket
	}
	if f.router == nil {
		f.router = NewRouter()
	}
	if f.exec == nil {
		f.exec = NewExecutor()
	}

	return f
}

type HandlerFilterOption func(*HandlerFilter)

type HandlerFilter struct {
	base.Filter
	frame   frame.IFrame
	codec   codec.ICodec
	creator PacketCreator
	router  IRouter
	exec    IExecutor
}

func (f *HandlerFilter) Name() string {
	return "HandlerFilter"
}

func (f *HandlerFilter) HandleRead(ctx anet.IFilterCtx) {
	buff, ok := ctx.Data().(*buffer.Buffer)
	if ok {
		ctx.Abort(nil)
		return
	}

	_, _ = buff.Seek(0, io.SeekStart)
	data, err := f.frame.Decode(buff)
	if err != nil {
		if err == frame.ErrIncomplete {
			err = nil
		}
		ctx.Abort(err)
		return
	}

	// 解析消息头,注意,此时并没有解析消息体,因为还不知具体消息体类型
	req := f.creator()
	if err := req.Decode(data); err != nil {
		ctx.Abort(err)
		return
	}

	req.SetCodec(f.codec)

	rsp := f.creator()
	nctx := NewContext(ctx.Conn(), req, rsp)
	if handler, err := f.router.Find(req); err != nil {
		ctx.Abort(err)
		return
	} else {
		nctx.SetHandler(handler)
	}

	if f.exec != nil {
		f.exec.Handle(nctx)
	} else {
		if err := nctx.Handler()(nctx); err != nil {
			ctx.Abort(err)
		}
	}
}

func (f *HandlerFilter) HandleWrite(ctx anet.IFilterCtx) {
	data := ctx.Data()

	// 如果data是buffer类型,则说明外部完全托管了消息序列化
	// 对于消息广播的情况:可以外边序列化好,将Buffer保存在Body中
	// IPacket:消息头每次都要单独序列化，因为每个人的消息头大概率都是不一样的
	pkg, ok := data.(IPacket)
	if !ok {
		ctx.Next()
		return
	}

	pkg.SetCodec(f.codec)

	buff := buffer.New()

	if err := pkg.Encode(buff); err != nil {
		ctx.Abort(err)
		return
	}

	// 写入Frame
	if err := f.frame.Encode(buff); err != nil {
		ctx.Abort(err)
		return
	}

	// 准备发送消息
	ctx.SetData(buff)
}

// Options
func WithFrame(f frame.IFrame) HandlerFilterOption {
	return func(opts *HandlerFilter) {
		opts.frame = f
	}
}

func WithCodec(c codec.ICodec) HandlerFilterOption {
	return func(opts *HandlerFilter) {
		opts.codec = c
	}
}

func WithCreator(c PacketCreator) HandlerFilterOption {
	return func(opts *HandlerFilter) {
		opts.creator = c
	}
}

func WithRouter(r IRouter) HandlerFilterOption {
	return func(opts *HandlerFilter) {
		opts.router = r
	}
}

func WithExecutor(e IExecutor) HandlerFilterOption {
	return func(opts *HandlerFilter) {
		opts.exec = e
	}
}
