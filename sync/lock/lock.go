package lock

import (
	"errors"
	"sync/atomic"
)

var ErrNotAcquire = errors.New("not acquire locker")

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
	Lock(key string, opts *LockOptions) (Unlocker, error)
}

type Unlocker interface {
	Unlock()
}

// 使用方法
// l, err := Lock("test_key")
// if err != nil {
//   return
// }
// defer l.Unlock()
func Lock(key string, opts ...LockOption) (Unlocker, error) {
	o := LockOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	if o.TTL <= 0 {
		o.TTL = DefaultTTL
	}

	return GetDefault().Lock(key, &o)
}
