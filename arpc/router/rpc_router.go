package router

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/util/times"
)

var (
	ErrInvalidSeqID    = errors.New("invalid sequence id")
	ErrInvalidResponse = errors.New("invalid response")
	ErrInvalidFuture   = errors.New("invalid future")
	ErrInvalidHandler  = errors.New("invalid handler")
	ErrNotSupport      = errors.New("not support")
	ErrHasStop         = errors.New("has stop")
)

func NewRPCRouter() arpc.RPCRouter {
	return &RpcRouter{ticker: time.Millisecond * 100}
}

type reply struct {
	Handler  arpc.Handler   // 消息处理函数
	SeqID    string         // 唯一RpcID
	Deadline int64          // 超时时间戳
	Future   arpc.Future    //
	RetryCB  arpc.RetryFunc // 重试回调
	RetryMax int            // 最大重试次数
	RetryNum int            // 当前重试次数
}

const (
	statusIdle = 0 // 未执行
	statusRun  = 1 // 正在执行
	statusStop = 2 // 已经停止
)

type RpcRouter struct {
	mux     sync.Mutex
	status  int
	ticker  time.Duration
	replies map[string]*reply
}

// Register 注册RPC回调,支持两种形式,Ptr同步阻塞调用,Func异步非阻塞调用
func (r *RpcRouter) Register(rsp interface{}, o *arpc.RegisterRPCOptions) error {
	if o.SeqID == "" {
		return ErrInvalidSeqID
	}

	v := reflect.ValueOf(rsp)
	t := v.Type()
	switch t.Kind() {
	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return ErrInvalidResponse
		}

		if o.Future == nil {
			return ErrInvalidFuture
		}

		handler := func(ctx arpc.IContext) error {
			if err := ctx.Response().Parse(rsp); err != nil {
				return err
			}
			return o.Future.Done()
		}
		return r.add(handler, o)
	case reflect.Func:
		// 传入的是函数指针,暗示非阻塞调用
		// 函数原型: func(IContext, *XXXResponse)
		if t.NumIn() != 2 || !arpc.IsContext(t.In(0)) {
			return ErrInvalidHandler
		}
		p1 := t.In(1)
		if p1.Kind() != reflect.Ptr || p1.Elem().Kind() != reflect.Struct {
			return ErrInvalidHandler
		}

		handler := func(ctx arpc.IContext) error {
			msg := reflect.New(t.In(1))
			if err := ctx.Response().Parse(msg.Interface()); err != nil {
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
		return ErrNotSupport
	}
}

func (r *RpcRouter) add(handler arpc.Handler, o *arpc.RegisterRPCOptions) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	deadline := times.Now() + int64(o.TTL/time.Millisecond)
	info := &reply{
		Handler:  handler,
		SeqID:    o.SeqID,
		Deadline: deadline,
		Future:   o.Future,
		RetryMax: o.RetryMax,
		RetryCB:  o.RetryCB,
		RetryNum: 0}
	r.replies[o.SeqID] = info

	switch r.status {
	case statusStop:
		return ErrHasStop
	case statusIdle:
		r.status = statusRun
		go r.Run()
	}

	return nil
}

func (r *RpcRouter) Find(seqId string) (arpc.Handler, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if reply, ok := r.replies[seqId]; ok {
		delete(r.replies, seqId)
		return reply.Handler, nil
	}

	return nil, arpc.ErrNoHandler
}

func (r *RpcRouter) Run() {
	t := time.NewTicker(r.ticker)
	for ; true; <-t.C {
		now := times.Now()
		needExit := false
		r.mux.Lock()
		needExit = r.status == statusStop
		// TODO:小顶堆优化?避免全部遍历同时还可以快速查询??
		for k, v := range r.replies {
			if v.Deadline >= now {
				v.RetryNum++
				delete(r.replies, k)
				if v.RetryCB != nil && v.RetryNum < v.RetryMax {
					seqID, ttl, err := v.RetryCB(v.RetryNum)
					if err == nil {
						v.SeqID = seqID
						v.Deadline = now + int64(time.Millisecond*ttl)
						r.replies[seqID] = v
					} else {
						_ = v.Future.Fail(err)
					}
				} else {
					_ = v.Future.Fail(arpc.ErrTimeout)
				}
			}
		}
		r.mux.Unlock()

		if needExit {
			break
		}
	}
}

func (r *RpcRouter) Stop() {
	r.mux.Lock()
	r.status = statusStop
	r.mux.Unlock()
}
