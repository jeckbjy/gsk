package arpc

import (
	"time"
)

func WithMsgID(msgid int) CallOption {
	return func(o *CallOptions) {
		o.ID = msgid
	}
}

func WithName(name string) CallOption {
	return func(o *CallOptions) {
		o.Name = name
	}
}

func WithMethod(m string) CallOption {
	return func(o *CallOptions) {
		o.Method = m
	}
}

func WithFuture(f Future) CallOption {
	return func(o *CallOptions) {
		o.Future = f
	}
}

func WithRetry(r int) CallOption {
	return func(o *CallOptions) {
		o.Retry = r
	}
}

func WithTTL(ttl time.Duration) CallOption {
	return func(o *CallOptions) {
		o.TTL = ttl
	}
}
