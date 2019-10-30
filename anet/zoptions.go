package anet

import "time"

type ListenOptions struct {
	Tag string
}

func (o *ListenOptions) Init(opts ...ListenOption) {
	for _, cb := range opts {
		cb(o)
	}
}

type DialOptions struct {
	Conn     Conn              // 老的连接,用于手动断线重连
	Tag      string            // 额外标识
	Timeout  time.Duration     // 超时时间,默认为0,表示不超时
	Blocking bool              // 是否阻塞,默认false
	RetryMax int               // 自动重连最大尝试次数,0:表示不重连,<0:表示一直尝试,默认为0不重连
	Interval time.Duration     // 每次重连间隔，默认1s
	Request  interface{}       // 连接成功后会立即发送一条消息,RetryMax为0
	Callback func(Conn, error) // 连接回调
}

func (o *DialOptions) Init(opts ...DialOption) {
	o.RetryMax = 0
	o.Interval = time.Second
	for _, cb := range opts {
		cb(o)
	}
}

type DialOption func(*DialOptions)
type ListenOption func(*ListenOptions)

func WithListenTag(tag string) ListenOption {
	return func(opts *ListenOptions) {
		opts.Tag = tag
	}
}

func WithDialTag(tag string) DialOption {
	return func(opts *DialOptions) {
		opts.Tag = tag
	}
}

func WithBlocking(blocking bool) DialOption {
	return func(opts *DialOptions) {
		opts.Blocking = blocking
	}
}

func WithTimeout(t time.Duration) DialOption {
	return func(opts *DialOptions) {
		opts.Timeout = t
	}
}

func WithRetryMax(t int) DialOption {
	return func(opts *DialOptions) {
		opts.RetryMax = t
	}
}

func WithInterval(t time.Duration) DialOption {
	return func(opts *DialOptions) {
		opts.Interval = t
	}
}

func WithCallback(fn func(Conn, error)) DialOption {
	return func(opts *DialOptions) {
		opts.Callback = fn
	}
}

func WithConn(conn Conn) DialOption {
	return func(opts *DialOptions) {
		opts.Conn = conn
	}
}
