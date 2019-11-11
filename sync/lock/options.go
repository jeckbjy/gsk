package lock

import (
	"time"
)

const (
	DefaultTTL   = time.Minute
	DefaultRetry = 1
)

type LockOptions struct {
	TTL   time.Duration // 锁过期时间,必须要设置,0使用默认
	Wait  time.Duration // 阻塞等待时间,默认为0不阻塞
	Retry int           // 重试次数,总的等待时间为wait*retry
}

type LockOption func(o *LockOptions)

// TTL sets the lock ttl
func TTL(t time.Duration) LockOption {
	return func(o *LockOptions) {
		o.TTL = t
	}
}

// Wait sets the wait time
func Wait(t time.Duration, retry int) LockOption {
	return func(o *LockOptions) {
		o.Wait = t
		o.Retry = retry
	}
}
