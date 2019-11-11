package lock

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

func newLocal() Locking {
	return &localLocking{locks: make(map[string]*localLocker)}
}

// 本地调试使用
// 不能用于分布式环境,不关心ttl过期,需要外边主动Unlock,才能被释放
type localLocking struct {
	mux   sync.Mutex
	locks map[string]*localLocker
}

func (l *localLocking) Name() string {
	return "local"
}

func (l *localLocking) Lock(key string, opts *LockOptions) (Unlocker, error) {
	if locker, err := l.tryLock(key); err == nil {
		return locker, nil
	}

	if opts != nil && opts.Wait > 0 {
		for retry := opts.Retry; retry > 0; retry-- {
			// sleep for retry
			time.Sleep(opts.Wait)
			if locker, err := l.tryLock(key); err == nil {
				return locker, nil
			}
		}
	}

	return nil, ErrNotAcquire
}

func (l *localLocking) tryLock(key string) (Unlocker, error) {
	l.mux.Lock()
	m, ok := l.locks[key]
	if !ok {
		m = &localLocker{owner: l, key: key}
		l.locks[key] = m
	}
	e := m.TryLock()
	l.mux.Unlock()

	if !e {
		return nil, ErrNotAcquire
	}

	return m, nil
}

func (l *localLocking) unlock(locker *localLocker) {
	l.mux.Lock()
	if _, ok := l.locks[locker.key]; ok {
		locker.mux.Unlock()
		delete(l.locks, locker.key)
	}
	l.mux.Unlock()
}

const mutexLocked = 1 << iota

type localLocker struct {
	owner *localLocking
	key   string
	mux   sync.Mutex
}

func (l *localLocker) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&l.mux)), 0, mutexLocked)
}

func (l *localLocker) Unlock() {
	l.owner.unlock(l)
}
