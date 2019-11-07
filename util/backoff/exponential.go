package backoff

import (
	"math/rand"
	"time"
)

var (
	defaultFactor   float64 = 2
	defaultMinDelay         = 100 * time.Millisecond
	defaultMaxDelay         = 2 * time.Second
)

func NewExponential(opts ...ExponentialOption) BackOff {
	b := &ExponentialBackOff{}

	for _, fn := range opts {
		fn(b)
	}
	if b.min <= 0 {
		b.min = defaultMinDelay
	}
	if b.max <= 0 {
		b.max = defaultMaxDelay
	}
	if b.factor <= 0 {
		b.factor = defaultFactor
	}

	return b
}

type ExponentialOption func(b *ExponentialBackOff)

func WithMin(d time.Duration) ExponentialOption {
	return func(b *ExponentialBackOff) {
		b.min = d
	}
}

func WithMax(d time.Duration) ExponentialOption {
	return func(b *ExponentialBackOff) {
		b.max = d
	}
}

func WithFactor(v float64) ExponentialOption {
	return func(b *ExponentialBackOff) {
		b.factor = v
	}
}

func WithJitter(v bool) ExponentialOption {
	return func(b *ExponentialBackOff) {
		b.jitter = v
	}
}

// https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
// https://www.baeldung.com/resilience4j-backoff-jitter
// https://github.com/grpc/grpc-go/blob/master/internal/backoff/backoff.go
// https://github.com/jpillora/backoff/blob/master/backoff.go
// https://github.com/rfyiamcool/backoff
//
// float64(b.Min) * math.Pow(b.Factor, b.Attempts)
// wait_interval = base * multiplier^n
type ExponentialBackOff struct {
	factor   float64
	jitter   bool
	min      time.Duration
	max      time.Duration
	attempts int64
	last     float64
}

func (e *ExponentialBackOff) Reset() {
	e.attempts = 0
	e.last = 0
}

func (e *ExponentialBackOff) Next() time.Duration {
	if e.last >= float64(e.max) {
		return e.max
	}

	var dur float64
	if e.last == 0 {
		dur = float64(e.min)
	} else {
		dur = e.last * e.factor
	}

	// cache
	e.last = dur

	if e.jitter {
		//[min, dur)
		dur = rand.Float64()*(dur-float64(e.min)) + float64(e.min)
	}

	if dur > float64(e.max) {
		dur = float64(e.max)
	}

	e.attempts++
	return time.Duration(dur)
}
