package breaker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func newBreaker(cfg *Config) Breaker {
	return &_CircuitBreaker{cfg: cfg, counter: newCounter(cfg.Window, cfg.Bucket, cfg.clock)}
}

// 参考:
// https://github.com/sony/gobreaker
// https://github.com/rubyist/circuitbreaker
// https://github.com/afex/hystrix-go
//
// https://github.com/Netflix/Hystrix/wiki/
// https://github.com/openbilibili/go-common
// https://zhuanlan.zhihu.com/p/58428026
// https://blog.csdn.net/tongtong_use/article/details/78611225
// https://segmentfault.com/a/1190000005988895
// https://techblog.constantcontact.com/software-development/circuit-breakers-and-microservices/
type _CircuitBreaker struct {
	cfg     *Config
	state   int32      // 当前状态
	counter *_Counter  // 滑动窗口,close状态下统计错误率
	expiry  time.Time  // open状态过期时间
	request int32      // half-open统计通过的流量
	mux     sync.Mutex //
}

func (b *_CircuitBreaker) Allow() error {
	state := State(atomic.LoadInt32(&b.state))
	if state == stateHealth || state == StateClosed {
		return nil
	}

	if b.cfg.Disabled {
		return nil
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	switch State(b.state) {
	case StateOpen:
		now := b.now()
		if now.After(b.expiry) {
			b.setState(StateHalfOpen)
			return nil
		}
		return ErrReject
	case StateHalfOpen:
		b.request++
		// TODO:通过配置允许多个?那么如果判断返回结果呢？任意成功或失败则修改状态?
		if b.request > 1 {
			return ErrTooMany
		}
		return nil
	default:
		return nil
	}
}

func (b *_CircuitBreaker) Mark(err error) {
	if err == nil {
		b.MarkSuccess()
	} else {
		b.MarkFailure()
	}
}

func (b *_CircuitBreaker) MarkSuccess() {
	// 健康状态不需要统计
	if b.getState() == stateHealth {
		return
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	switch State(b.state) {
	case StateClosed:
		b.counter.AddSuccess()
		total := b.counter.Total()
		if total.Total() >= int64(b.cfg.RequestThreshold) && b.counter.ErrorRate() == 0 {
			// 进入快速模式,不再统计数据
			b.setState(stateHealth)
			b.counter.Reset()
		}
	case StateHalfOpen:
		b.setState(StateClosed)
		b.counter.Reset()
		b.counter.AddSuccess()
	default:
		// open or Health state,ignore, do nothing
	}
}

func (b *_CircuitBreaker) MarkFailure() {
	b.mux.Lock()
	defer b.mux.Unlock()
	switch State(b.state) {
	case stateHealth:
		// reset data
		b.setState(StateClosed)
		b.counter.Reset()
		b.counter.AddFailure()
	case StateClosed:
		b.counter.AddFailure()
		total := b.counter.Total()
		if total.Total() > int64(b.cfg.RequestThreshold) && total.ErrorRate() > b.cfg.Ratio {
			b.setState(StateOpen)
			b.expiry = b.now().Add(b.cfg.Sleep)
		}
	case StateHalfOpen:
		b.setState(StateOpen)
		b.expiry = b.now().Add(b.cfg.Sleep)
	default:
		// open,ignore,do nothing
	}
}

func (b *_CircuitBreaker) String() string {
	total := b.counter.Total()
	return fmt.Sprintf(
		"state %+v,success %d, failture %d, error_rate %f",
		State(b.state),
		total.success,
		total.failure,
		total.ErrorRate())
}

func (b *_CircuitBreaker) getState() State {
	return State(atomic.LoadInt32(&b.state))
}

func (b *_CircuitBreaker) setState(s State) {
	atomic.StoreInt32(&b.state, int32(s))
}

func (b *_CircuitBreaker) now() time.Time {
	return b.cfg.clock.Now()
}
