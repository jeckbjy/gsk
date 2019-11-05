package breaker

import (
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	c := newCounter(defaultWindow, defaultBucket, newMockClock(defaultWindow, time.Now(), true))
	for i := 0; i < 100; i++ {
		c.AddSuccess()
	}
	// c.Successes() == defaultBucket
	// 10
	t.Log(c.Successes())

	c = newCounter(defaultWindow, defaultBucket, newMockClock(defaultWindow/2, time.Now(), true))
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			c.AddSuccess()
		} else {
			c.AddFailure()
		}
	}

	// 20 0.5
	t.Log(c.Successes(), c.ErrorRate())
}

func TestGroup(t *testing.T) {
	b1 := Get("key1")
	b2 := Get("key2")
	if b1 == b2 {
		t.FailNow()
	}

	b3 := Get("key1")
	if b1 != b3 {
		t.FailNow()
	}
}

func TestBreaker(t *testing.T) {
	clock := newMockClock(defaultWindow, time.Now(), false)
	cfg := &Config{}
	cfg.clock = clock
	cfg.fix()

	//
	b1 := newBreaker(cfg)
	markFailure(b1, 20)
	markSuccess(b1, 80)
	if b1.Allow() == nil {
		t.Log(b1.String())
	} else {
		t.FailNow()
	}

	// 熔断
	b2 := newBreaker(cfg)
	// 超过默认值并失败率(1)大于配置值
	markFailure(b2, 21)
	if b2.Allow() != nil {
		t.Log(b2.String())
	} else {
		t.FailNow()
	}
	// 超过静默时间,进入half-open状态,允许一个通过
	clock.Add(defaultSleep * 2)
	if b2.Allow() != nil {
		t.Log(b2.String())
		t.FailNow()
	} else {
		t.Log(b2.String())
	}
	// half-open ===> closed
	b2.MarkSuccess()
	if b2.Allow() != nil {
		t.Log(b2.String())
		t.FailNow()
	} else {
		t.Log(b2.String())
	}
}

func markSuccess(b Breaker, num int) {
	for i := 0; i < num; i++ {
		b.MarkSuccess()
	}
}

func markFailure(b Breaker, num int) {
	for i := 0; i < num; i++ {
		b.MarkFailure()
	}
}
