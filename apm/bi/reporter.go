package bi

import (
	"sync"
	"time"

	"github.com/jeckbjy/gsk/util/idgen/xid"
)

const (
	statusNone = 0
	statusRun  = 1
	statusStop = 2
)

const (
	defaultMax   = 100000
	defaultBatch = 100
	defaultWait  = time.Minute * 10
)

// 配置信息,通常只需要设置URL,其他使用默认值
type Options struct {
	Tran   Transport     // 默认http
	URL    string        // 发送地址
	Batch  int           // 批量发送大小
	Max    int           // 最大缓存数,超出则丢弃
	Wait   time.Duration // 最大等待时间,超过此时间,则强制退出
	Encode Encode        // 编码器
}

// 编码器,默认json编码,并且自动名固定
type Encode func(events []*Message) []byte

type M = map[string]interface{}

type Message struct {
	ID     string
	Time   time.Time
	Event  string
	Params M
	next   *Message
}

// 用于编码并发送消息
type Transport interface {
	Send(opts *Options, events []byte) error
}

// BI统计信息
// 每条BI数据要求有一个全局唯一ID可用于排重
// 要有一个唯一事件ID,用于区分参数含义
// 时间戳,标识发送时间
// 任意kv参数
//
// message不需要发送有序,因为ID自增有序,且有时间戳
// bi数据比较重要,需要失败重传,需要能缓存一定数量数据,可Batch发送
//
// 默认只实现了http协议,效率可能没有那么高
// 为了避免消息积压以及稳定,可以考虑使用MQ
type Reporter struct {
	queue   Queue
	mux     sync.Mutex
	cond    *sync.Cond
	opts    *Options
	status  int
	expired time.Time // 超时则强制退出
	success int       // 统计成功发送次数
	fail    int       // 统计失败发送次数
	exit    chan bool
}

func (r *Reporter) Init(opts *Options) error {
	if opts.Tran == nil && opts.URL == "" {
		return ErrBadOption
	}

	if opts.Tran == nil {
		opts.Tran = NewHttp()
	}

	if opts.Encode == nil {
		opts.Encode = DefaultEncode
	}
	if opts.Max == 0 {
		opts.Max = defaultMax
	}
	if opts.Batch == 0 {
		opts.Batch = defaultBatch
	}
	if opts.Wait == 0 {
		opts.Wait = defaultWait
	}
	if r.status != statusNone {
		return ErrHasInit
	}

	r.mux.Lock()
	r.opts = opts
	r.status = statusRun
	r.mux.Unlock()
	go r.run()
	return nil
}

func (r *Reporter) Stop() {
	r.mux.Lock()
	if r.status != statusRun {
		r.mux.Unlock()
		return
	}
	r.status = statusStop
	r.mux.Unlock()
	r.cond.Signal()
	// wait for exit
	<-r.exit
}

func (r *Reporter) Send(event string, params M) error {
	r.mux.Lock()
	if r.opts.Tran == nil {
		r.mux.Unlock()
		return ErrNotInit
	}

	if r.queue.Len() > r.opts.Max {
		r.mux.Unlock()
		return ErrTooManyMsg
	}

	msg := &Message{
		ID:     xid.New().String(),
		Time:   time.Now(),
		Event:  event,
		Params: params,
	}

	r.queue.Push(msg)
	r.mux.Unlock()
	r.cond.Signal()

	return nil
}

func (r *Reporter) run() {
	records := make([]*Message, r.opts.Batch)
	for {
		r.mux.Lock()
		for r.status != statusStop && r.queue.Empty() {
			r.cond.Wait()
		}
		batch := r.queue.Pop(records)
		quit := r.canQuit()
		r.mux.Unlock()
		r.doSend(batch)
		if quit {
			break
		}
	}

	r.exit <- true
}

func (r *Reporter) doSend(batch []*Message) {
	if len(batch) == 0 {
		return
	}

	tran := r.opts.Tran
	data := r.opts.Encode(batch)
	// 一直尝试发送
	for {
		if tran.Send(r.opts, data) == nil {
			// 成功
			r.success++
			break
		} else {
			r.fail++
		}
		r.mux.Lock()
		quit := r.canQuit()
		r.mux.Unlock()
		if quit {
			break
		}
	}
}

func (r *Reporter) canQuit() bool {
	if r.status == statusRun {
		return false
	}

	if r.queue.Empty() {
		return true
	}

	if r.success == 0 {
		return true
	}

	if time.Now().After(r.expired) {
		return true
	}

	return false
}
