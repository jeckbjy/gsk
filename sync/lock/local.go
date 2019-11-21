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

func (l *localLocking) Acquire(key string, opts *Options) (Locker, error) {
	locker := &localLocker{owner: l, key: key, opts: opts}
	return locker, nil
}

type localLocker struct {
	owner *localLocking
	key   string
	opts  *Options
}

func (l *localLocker) Lock() error {
	mux, locked := l.tryLock()
	if locked {
		return nil
	}

	timeout := time.Duration(0)
	if l.opts != nil {
		timeout = l.opts.Timeout
	}

	if timeout == TimeoutMax {
		// 直到获取到锁
		mux.Lock()
		return nil
	} else {
		// 轮询检测是否获得到锁,直到过期
		expired := time.Now().Add(timeout)
		for {
			time.Sleep(time.Millisecond)
			if _, locked := l.tryLock(); locked {
				return nil
			}
			if time.Now().After(expired) {
				return ErrNotLock
			}
		}
	}
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

func (l *localLocker) tryLock() (*sync.Mutex, bool) {
	p := l.owner
	p.mux.Lock()
	mux, ok := p.locks[l.key]
	if !ok {
		mux = &sync.Mutex{}
		p.locks[l.key] = mux
	}
	locked := tryLock(mux)
	p.mux.Unlock()
	return mux, locked
}

const mutexLocked = 1

func tryLock(mux *sync.Mutex) bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(mux)), 0, mutexLocked)
}
