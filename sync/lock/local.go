package lock

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

func newLocal() Locking {
	return &localLocking{locks: make(map[string]*sync.Mutex)}
}

// 本地调试使用
// 不能用于分布式环境,不关心ttl过期,需要外边主动Unlock,才能被释放
type localLocking struct {
	mux   sync.Mutex
	locks map[string]*sync.Mutex
}

func (l *localLocking) Name() string {
	return "local"
}

func (l *localLocking) Acquire(key string, opts *LockOptions) (Locker, error) {
	locker := &localLocker{owner: l, key: key, opts: opts}
	return locker, nil
}

type localLocker struct {
	owner *localLocking
	key   string
	opts  *LockOptions
}

func (l *localLocker) Lock() error {
	if l.tryLock() {
		return nil
	}

	opts := l.opts
	if opts != nil && opts.Wait > 0 {
		for retry := opts.Retry; retry > 0; retry-- {
			time.Sleep(opts.Wait)
			if l.tryLock() {
				return nil
			}
		}
	}

	return ErrNotLock
}

func (l *localLocker) Unlock() {
	p := l.owner
	p.mux.Lock()
	if mux, ok := p.locks[l.key]; ok {
		mux.Unlock()
		delete(p.locks, l.key)
	}
	p.mux.Unlock()
}

func (l *localLocker) tryLock() bool {
	p := l.owner
	p.mux.Lock()
	mux, ok := p.locks[l.key]
	if !ok {
		mux = &sync.Mutex{}
		p.locks[l.key] = mux
	}
	locked := tryLock(mux)
	p.mux.Unlock()
	return locked
}

const mutexLocked = 1

func tryLock(mux *sync.Mutex) bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(mux)), 0, mutexLocked)
}
