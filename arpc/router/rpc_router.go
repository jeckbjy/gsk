package router

import (
	"reflect"
	"sync"
	"time"

	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/util/timex"
)

type _RpcInfo struct {
	Handler arpc.HandlerFunc // 消息回调
	Request arpc.Packet      // 发送的请求
	Data    interface{}      // 需要透传的数据
	Timer   *timex.Timer     // 定时器
	Retry   arpc.RetryFunc   // 重试函数
}

type RpcRouter struct {
	mux   sync.Mutex
	infos map[string]*_RpcInfo
}

func (r *RpcRouter) Init() {
	r.infos = make(map[string]*_RpcInfo)
}

func (r *RpcRouter) Handle(ctx arpc.Context) arpc.HandlerFunc {
	pkg := ctx.Message()
	r.mux.Lock()
	info, ok := r.infos[pkg.SeqID()]
	if ok {
		info.Timer.Stop()
		delete(r.infos, pkg.SeqID())
	}
	r.mux.Unlock()

	if info != nil {
		ctx.SetData(info.Data)
		return info.Handler
	} else { // 不存在也不需要报错
		return nil
	}
}

func (r *RpcRouter) Register(request arpc.Packet) error {
	opts := request.Internal().(*arpc.MiscOptions)
	t := reflect.TypeOf(opts.Response)
	switch t.Kind() {
	case reflect.Ptr:
		return r.registerSync(request, opts, t)
	case reflect.Func:
		return r.registerAsync(request, opts)
	default:
		return arpc.ErrNotSupport
	}
}

// 同步调用,必须有Future
func (r *RpcRouter) registerSync(req arpc.Packet, opts *arpc.MiscOptions, t reflect.Type) error {
	// 指针类型,阻塞同步回调
	if t.Elem().Kind() != reflect.Struct {
		return arpc.ErrInvalidResponse
	}

	if opts.Future == nil {
		return arpc.ErrInvalidFuture
	}

	handler := func(ctx arpc.Context) error {
		err := arpc.DecodeBody(ctx.Message(), opts.Response)
		opts.Future.Done(err)
		return err
	}

	return r.add(req, opts, handler)
}

// 异步调用,没必要使用Future,故必须为nil
// 支持的原型有
// 原型1: func(ctx Context) error
// 原型2: func(rsp *Response) error
// 原型3: func(ctx Context, rsp *Response) error
func (r *RpcRouter) registerAsync(request arpc.Packet, opts *arpc.MiscOptions) error {
	if opts.Future != nil {
		return arpc.ErrInvalidFuture
	}

	// 函数原型,完全需要用户自己处理
	if handler, ok := opts.Response.(arpc.HandlerFunc); ok {
		return r.add(request, opts, handler)
	}

	v := reflect.ValueOf(opts.Response)
	t := v.Type()

	switch t.NumIn() {
	case 1: // func(rsp *Response) error
		if !isMessage(t.In(0)) || t.NumOut() != 1 || !isError(t.Out(0)) {
			return arpc.ErrInvalidHandler
		}

		handler := func(ctx arpc.Context) error {
			msg := ctx.Message()
			if msg.Body() == nil {
				if err := arpc.DecodeBody(msg, reflect.New(t.In(0).Elem()).Interface()); err != nil {
					return err
				}
			}

			in := []reflect.Value{reflect.ValueOf(msg.Body())}
			out := v.Call(in)
			if !out[0].IsNil() {
				err := out[0].Interface().(error)
				return err
			}
			return nil
		}

		return r.add(request, opts, handler)
	case 2: // func(ctx Context, rsp *Response) error
		if !isContext(t.In(0)) || !isMessage(t.In(1)) || t.NumOut() != 1 || !isError(t.Out(0)) {
			return arpc.ErrInvalidHandler
		}

		handler := func(ctx arpc.Context) error {
			msg := ctx.Message()
			if msg.Body() == nil {
				if err := arpc.DecodeBody(msg, reflect.New(t.In(1).Elem()).Interface()); err != nil {
					return err
				}
			}

			in := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg.Body())}
			out := v.Call(in)
			if !out[0].IsNil() {
				err := out[0].Interface().(error)
				return err
			}
			return nil
		}

		return r.add(request, opts, handler)
	default:
		return arpc.ErrInvalidHandler
	}
}

func (r *RpcRouter) add(req arpc.Packet, opts *arpc.MiscOptions, handler arpc.HandlerFunc) error {
	r.mux.Lock()
	if opts.Future != nil {
		opts.Future.Add()
	}

	info := &_RpcInfo{Handler: handler, Request: req, Data: opts.Extra}
	r.infos[req.SeqID()] = info
	timer := timex.NewDelayTimer(opts.TTL, func() {
		r.onTimeout(req.SeqID())
	})
	info.Timer = timer

	r.mux.Unlock()
	return nil
}

func (r *RpcRouter) onTimeout(seqID string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if info, ok := r.infos[seqID]; ok {
		ttl := time.Duration(0)
		if info.Retry != nil {
			ttl = info.Retry(info.Request)
		}

		if ttl > 0 {
			seqID := info.Request.SeqID()
			info.Timer = timex.NewDelayTimer(ttl, func() {
				r.onTimeout(seqID)
			})
		} else {
			delete(r.infos, seqID)
			// TODO:notify timeout for async callback, add onError callback?
			opts := info.Request.Internal().(*arpc.MiscOptions)
			if opts.Future != nil {
				opts.Future.Done(arpc.ErrTimeout)
			}
		}
	}
}
