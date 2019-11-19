package router

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/jeckbjy/gsk/arpc"
)

var (
	ErrInvalidSeqID    = errors.New("invalid sequence id")
	ErrInvalidResponse = errors.New("invalid response")
	ErrInvalidFuture   = errors.New("invalid future")
	ErrInvalidHandler  = errors.New("invalid handler")
	ErrNotSupport      = errors.New("not support")
	ErrHasStopped      = errors.New("has stopped")
)

const (
	statusIdle = 0
	statusRun  = 1
	statusStop = 2
)

type _RpcInfo struct {
	Handler  arpc.Handler
	SeqID    string         // RPC回调ID
	Expire   int64          // 过期时间
	Future   arpc.Future    // 完成通知
	RetryCB  arpc.RetryFunc // 重试回调
	RetryMax int            // 最大重试次数
	RetryNum int            // 当前重试次数
	Req      arpc.Packet
}

type _RpcRouter struct {
	infos  map[string]*_RpcInfo
	status int
	mux    sync.Mutex
}

func (r *_RpcRouter) Init() {
	r.infos = make(map[string]*_RpcInfo)
}

func (r *_RpcRouter) Find(pkg arpc.Packet) (arpc.Handler, error) {
	// find and delete
	r.mux.Lock()
	i, ok := r.infos[pkg.SeqID()]
	if ok {
		delete(r.infos, pkg.SeqID())
	}
	r.mux.Unlock()
	if i != nil && i.Handler != nil {
		return i.Handler, nil
	}

	return nil, arpc.ErrNoHandler
}

// 注册RPC回调,支持两种形式,Ptr同步阻塞调用,Func异步非阻塞调用
func (r *_RpcRouter) Register(rsp interface{}, o *arpc.RegisterOptions) error {
	v := reflect.ValueOf(rsp)
	t := v.Type()
	switch t.Kind() {
	case reflect.Ptr:
		// 阻塞同步回调
		if t.Elem().Kind() != reflect.Struct {
			return ErrInvalidResponse
		}

		if o.Future == nil {
			return ErrInvalidFuture
		}

		handler := func(ctx arpc.Context) error {
			if err := ctx.Response().DecodeBody(rsp); err != nil {
				return err
			}
			return o.Future.Done()
		}
		return r.add(handler, o)
	case reflect.Func:
		// 传入的是函数指针,暗示非阻塞调用
		// 函数原型1:
		// func(*XXXResponse)
		// func(Context, *XXXResponse)
		switch t.NumIn() {
		case 1:
			if arpc.IsContext(t.In(0)) {
				return ErrInvalidHandler
			}
			p0 := t.In(0)
			if p0.Kind() != reflect.Ptr || p0.Elem().Kind() != reflect.Struct {
				return ErrInvalidHandler
			}
			handler := func(ctx arpc.Context) error {
				msg := reflect.New(p0)
				if err := ctx.Response().DecodeBody(msg.Interface()); err != nil {
					return err
				}
				in := []reflect.Value{msg}
				v.Call(in)
				if o.Future != nil {
					return o.Future.Done()
				}

				return nil
			}
			return r.add(handler, o)
		case 2:
			if !arpc.IsContext(t.In(0)) {
				return ErrInvalidHandler
			}

			p1 := t.In(1)
			if p1.Kind() != reflect.Ptr || p1.Elem().Kind() != reflect.Struct {
				return ErrInvalidHandler
			}

			handler := func(ctx arpc.Context) error {
				msg := reflect.New(t.In(1))
				if err := ctx.Response().DecodeBody(msg.Interface()); err != nil {
					return err
				}

				in := []reflect.Value{reflect.ValueOf(ctx), msg}
				v.Call(in)
				if o.Future != nil {
					return o.Future.Done()
				}

				return nil
			}

			return r.add(handler, o)
		default:
			return ErrInvalidHandler
		}

	default:
		return ErrNotSupport
	}
}

func (r *_RpcRouter) add(handler arpc.Handler, o *arpc.RegisterOptions) error {
	var err error
	r.mux.Lock()
	expired := time.Now().Add(o.TTL).UnixNano() / int64(time.Millisecond)
	info := &_RpcInfo{
		Handler:  handler,
		SeqID:    o.SeqID,
		Expire:   expired,
		Future:   o.Future,
		RetryCB:  o.RetryCB,
		RetryMax: o.RetryMax,
		RetryNum: 0,
		Req:      o.Req,
	}
	r.infos[o.SeqID] = info
	switch r.status {
	case statusStop:
		err = ErrHasStopped
	case statusIdle:
		r.status = statusRun
		go r.Run()
	}
	r.mux.Unlock()
	return err
}

// TODO:检测过期
func (r *_RpcRouter) Run() {

}

func (r *_RpcRouter) Close() {
	r.mux.Lock()
	r.status = statusStop
	r.mux.Unlock()
}

//func (r *RpcRouter) Run() {
//	t := time.NewTicker(r.ticker)
//	for ; true; <-t.C {
//		now := times.Now()
//		needExit := false
//		r.mux.Lock()
//		needExit = r.status == statusStop
//		// TODO:小顶堆优化?避免全部遍历同时还可以快速查询??
//		for k, v := range r.replies {
//			if v.Deadline >= now {
//				v.RetryNum++
//				delete(r.replies, k)
//				if v.RetryCB != nil && v.RetryNum < v.RetryMax {
//					seqID, ttl, err := v.RetryCB(v.RetryNum)
//					if err == nil {
//						v.SeqID = seqID
//						v.Deadline = now + int64(time.Millisecond*ttl)
//						r.replies[seqID] = v
//					} else {
//						_ = v.Future.Fail(err)
//					}
//				} else {
//					_ = v.Future.Fail(arpc.ErrTimeout)
//				}
//			}
//		}
//		r.mux.Unlock()
//
//		if needExit {
//			break
//		}
//	}
//}
//
