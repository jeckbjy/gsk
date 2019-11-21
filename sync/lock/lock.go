package lock

import (
	"errors"
	"sync/atomic"
)

var ErrNotLock = errors.New("not lock")

var locking atomic.Value

func init() {
	SetDefault(newLocal())
}

func GetDefault() Locking {
	return locking.Load().(Locking)
}

func SetDefault(l Locking) {
	locking.Store(l)
}

// 分布式锁系统,distributed locking system
// 可以基于redis实现,也可以基于consul,etcd等实现
//
// 注意:因为目标是分布式锁,为了防止永久死锁,必须要设置一个过期,插件库来保证自动续期
//
// github.com/go-redsync/redsync
type Locking interface {
	Name() string
	Acquire(key string, opts *LockOptions) (Locker, error)
}

type Locker interface {
	Lock() error
	Unlock()
}

// 使用方法
// l, err := Lock("test_key")
// if err != nil {
//   return
// }
// defer l.Unlock()
func Lock(key string, opts ...LockOption) (Locker, error) {
	o := LockOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	if o.TTL <= 0 {
		o.TTL = DefaultTTL
	}

	l, err := GetDefault().Acquire(key, &o)
	if err != nil {
		return nil, err
	}
	if err := l.Lock(); err != nil {
		return nil, err
	}
	return l, nil
}
