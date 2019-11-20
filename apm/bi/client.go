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
	defaultRetry = 10
	defaultWait  = time.Minute * 10
)

// 配置信息,通常只需要设置URL,其他使用默认值
type Options struct {
	Tran   Transport     // 默认http
	URL    string        // 发送地址
	Batch  int           // 批量发送大小
	Retry  int           // 尝试重发次数
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

type Stat struct {
	Success int   // 成功次数
	Fail    int   // 失败次数
	Err     error // 最后一次错误
}

// BI统计
// 一:需要统计的数据
// 全局ID:可用于重发排重
// 事件ID:用于区分参数含义
// 时间戳:标识发送时间
// 参数:  任意kv结构
//
// 二:支持的特性
// 1.异步发送
// 2.缓存一定数量,超过则会丢失
// 3.可批量发送
// 4.json编码
// 5.默认http发送,可扩展
//
// 三:一些注意
// 1.虽然会缓存数据,但是超过上限后依然会丢弃数据,防止占用内存过大导致宕机
// 2.默认实现了http协议,效率可能没有那么高,可考虑使用MQ
type Client struct {
	queue   Queue
	mux     sync.Mutex
	cond    *sync.Cond
	opts    *Options
	status  int
	expired time.Time // 超时则强制退出
	stat    Stat      // 统计数据
	exit    chan bool
}

func (r *Client) Stat() Stat {
	return r.stat
}

func (r *Client) Init(opts *Options) error {
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
	if opts.Retry == 0 {
		opts.Retry = defaultRetry
	}
	if r.status != statusNone {
		return ErrHasInit
	}

	r.mux.Lock()
	r.cond = sync.NewCond(&r.mux)
	r.opts = opts
	r.status = statusRun
	r.exit = make(chan bool)
	r.mux.Unlock()
	go r.run()
	return nil
}

func (r *Client) Stop() {
	r.mux.Lock()
	if r.status != statusRun {
		r.mux.Unlock()
		return
	}
	r.status = statusStop
	r.expired = time.Now().Add(r.opts.Wait)
	r.mux.Unlock()
	r.cond.Signal()
	// wait for exit
	<-r.exit
}

func (r *Client) Send(event string, params M) error {
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

func (r *Client) run() {
	records := make([]*Message, r.opts.Batch)
	for {
		r.mux.Lock()
		for r.status != statusStop && r.queue.Empty() {
			r.cond.Wait()
		}
		batch := r.queue.Pop(records)
		quit := r.canQuit() || r.queue.Empty()
		r.mux.Unlock()
		if len(batch) > 0 {
			quit = r.doSend(batch)
		}

		if quit {
			break
		}
	}

	r.exit <- true
}

func (r *Client) doSend(batch []*Message) bool {
	tran := r.opts.Tran
	data := r.opts.Encode(batch)
	retry := r.opts.Retry
	// 一直尝试发送
	count := 0
	for {
		if err := tran.Send(r.opts, data); err == nil {
			// 成功
			r.stat.Success++
			break
		} else {
			count++
			r.stat.Fail++
			r.stat.Err = err
		}

		if count >= retry {
			// 判断是否需要退出
			r.mux.Lock()
			quit := r.canQuit()
			r.mux.Unlock()
			if quit {
				return true
			}
		}
	}

	return false
}

func (r *Client) canQuit() bool {
	if r.status == statusRun {
		return false
	}

	if r.stat.Success == 0 {
		return true
	}

	if time.Now().After(r.expired) {
		return true
	}

	return false
}
