package backoff

import "time"

type ZeroBackOff struct {
}

func (b *ZeroBackOff) Reset() {
}

func (b *ZeroBackOff) Next() time.Duration {
	return 0
}

type StopBackOff struct {
}

func (b *StopBackOff) Reset() {
}

func (b *StopBackOff) Next() time.Duration {
	return Stop
}

type ConstantBackOff struct {
	Interval time.Duration
}

func (b *ConstantBackOff) Reset() {
}

func (b *ConstantBackOff) Next() time.Duration {
	return b.Interval
}

func NewConstant(d time.Duration) *ConstantBackOff {
	return &ConstantBackOff{Interval: d}
}
