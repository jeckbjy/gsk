package fexec

import (
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/util/buffer"
)

// 创建execFilter,可以调用Router,Executor,Middleware覆盖默认参数
func New(opts ...Option) anet.Filter {
	f := &execFilter{
		router:   arpc.DefaultRouter(),
		executor: exec.Default(),
	}

	for _, fn := range opts {
		fn(f)
	}

	f.pool.New = func() interface{} {
		return &Task{pool: &f.pool}
	}
	return f
}

type Option func(f *execFilter)

func Router(router arpc.Router) Option {
	return func(f *execFilter) {
		if router != nil {
			f.router = router
		}
	}
}

func Executor(executor exec.Executor) Option {
	return func(f *execFilter) {
		if executor != nil {
			f.executor = executor
		}
	}
}

func Middleware(m ...arpc.Middleware) Option {
	return func(f *execFilter) {
		f.middleware = append(f.middleware, m...)
	}
}

func Middlewares(m []arpc.Middleware) Option {
	return func(f *execFilter) {
		f.middleware = append(f.middleware, m...)
	}
}

// execFilter 用于注册消息回调,或执行回调
type execFilter struct {
	base.Filter
	router     arpc.Router
	executor   exec.Executor
	middleware []arpc.Middleware
	pool       sync.Pool
}

func (f *execFilter) Name() string {
	return "exec"
}

func (f *execFilter) HandleRead(ctx anet.FilterCtx) error {
	data, ok := ctx.Data().(*buffer.Buffer)
	if !ok {
		return nil
	}

	msg := arpc.NewPacket()
	if err := msg.Decode(data); err != nil {
		return err
	}

	handler, err := f.router.Find(msg)
	if err != nil {
		return err
	}

	taskCtx := arpc.NewContext()
	taskCtx.Init(ctx.Conn(), msg)
	task := f.pool.Get().(*Task)
	task.Init(taskCtx, handler, f.middleware)
	return f.executor.Post(task)
}

func (f *execFilter) HandleWrite(ctx anet.FilterCtx) error {
	if pkt, ok := ctx.Data().(arpc.Packet); ok {
		buff := buffer.New()
		if err := pkt.Encode(buff); err != nil {
			return err
		}

		// 注册到router中监听回调和超时
		if info := pkt.CallInfo(); info != nil {
			if err := f.router.RegisterRPC(pkt); err != nil {
				return err
			}
		}

		ctx.SetData(buff)
	}

	return nil
}
