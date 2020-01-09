package fexec

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/util/buffer"
)

// 创建execFilter,可以调用Router,Executor,Middleware覆盖默认参数
func New(opts ...Option) anet.Filter {
	f := &execFilter{
		router:   arpc.GetRouter(),
		executor: exec.Default(),
	}

	for _, fn := range opts {
		fn(f)
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

// execFilter 用于注册消息回调,或执行回调
type execFilter struct {
	base.Filter
	router   arpc.Router
	executor exec.Executor
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

	taskCtx := arpc.NewContext()
	taskCtx.Init(ctx.Conn(), msg)
	task := newTask(taskCtx, f.router)
	task.Init(taskCtx, f.router)

	return f.executor.Post(task)
}

func (f *execFilter) HandleWrite(ctx anet.FilterCtx) error {
	if pkt, ok := ctx.Data().(arpc.Packet); ok {
		buff := buffer.New()
		if err := pkt.Encode(buff); err != nil {
			return err
		}

		// 注册到router中监听回调和超时
		if _, ok := pkt.Internal().(*arpc.MiscOptions); ok {
			if err := f.router.Register(pkt); err != nil {
				return err
			}
		}

		ctx.SetData(buff)
	}

	return nil
}
