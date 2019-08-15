package internal

import "time"

type SelectCB func(sk *SelectionKey)

type SelectOptions struct {
	Timeout  int      // 毫秒,-1表示永久
	Callback SelectCB // 设置回调,将会直接调用回调函数,而不会返回[]*SelectionKey数组
}

type SelectOption func(*SelectOptions)

func WithTimeout(t time.Duration) SelectOption {
	return func(o *SelectOptions) {
		o.Timeout = int(t / time.Millisecond)
	}
}

func WithCallback(cb SelectCB) SelectOption {
	return func(o *SelectOptions) {
		o.Callback = cb
	}
}
