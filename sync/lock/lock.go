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

// New 使用默认的locking创建一个locker
func New(key string, opts ...Option) (Locker, error) {
	o := Options{}
	o.Build(opts...)
	return GetDefault().Acquire(key, &o)
}

// New and Lock
// l, err := Lock("test_key")
// if err != nil {
//   return
// }
// defer l.Unlock()
func Lock(key string, opts ...Option) (Locker, error) {
	o := Options{}
	o.Build(opts...)
	l, err := GetDefault().Acquire(key, &o)
	if err != nil {
		return nil, err
	}
	if err := l.Lock(); err != nil {
		return nil, err
	}
	return l, nil
}

// Locking 分布式锁系统,distributed locking system
// 可以基于redis实现,也可以基于consul,etcd等实现
// 可选参数:
// TTL:通常不需要设置,用于指示业务过期时间,防止服务器宕机永远无法释放锁
// Timeout:用于指示Lock等待超时时间,0立即返回,-1表示永久等待,直到获取到锁
//
// 注意:因为目标是分布式锁,为了防止永久死锁,必须要设置一个过期,插件库来保证自动续期
// TODO:如何抽象乐观锁?
//
// github.com/go-redsync/redsync
type Locking interface {
	Name() string
	Acquire(key string, opts *Options) (Locker, error)
}

// Locker 类似于sync.Locker，但是Lock会返回错误
type Locker interface {
	Lock() error
	Unlock()
}
