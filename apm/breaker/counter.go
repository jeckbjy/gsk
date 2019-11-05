package breaker

import (
	"sync"
	"time"
)

func newCounter(window time.Duration, count int, clock _Clock) *_Counter {
	c := &_Counter{}
	c.mux.Lock()

	c.total.Reset()
	c.window = window
	c.buckets = make([]*_Bucket, count)
	c.offset = 0
	c.lastAccess = time.Time{}
	c.clock = clock

	for i := 0; i < count; i++ {
		c.buckets[i] = &_Bucket{}
	}

	c.mux.Unlock()
	return c
}

type _Bucket struct {
	success int64
	failure int64
}

func (b *_Bucket) Reset() {
	b.success = 0
	b.failure = 0
}

func (b *_Bucket) Total() int64 {
	return b.success + b.failure
}

func (b *_Bucket) ErrorRate() float64 {
	if b.failure == 0 {
		return 0
	}
	return float64(b.failure) / float64(b.success+b.failure)
}

type _Counter struct {
	total      _Bucket       // cache total
	buckets    []*_Bucket    // ring buffer
	offset     int           // ring offset
	lastAccess time.Time     // last access time
	window     time.Duration //
	clock      _Clock        // mock time.Now()
	mux        sync.RWMutex
}

func (c *_Counter) ErrorRate() float64 {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.total.ErrorRate()
}

func (c *_Counter) Successes() int64 {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.total.success
}

func (c *_Counter) Failures() int64 {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.total.failure
}

func (c *_Counter) Requests() int64 {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.total.success + c.total.failure
}

func (c *_Counter) Total() _Bucket {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.total
}

func (c *_Counter) AddSuccess() {
	c.mux.Lock()
	defer c.mux.Unlock()
	l := c.getLatestBucket()
	l.success++
	c.total.success++
}

func (c *_Counter) AddFailure() {
	c.mux.Lock()
	defer c.mux.Unlock()
	l := c.getLatestBucket()
	l.failure++
	c.total.failure++
}

func (c *_Counter) Reset() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.resetAllBuckets()
}

func (c *_Counter) getLatestBucket() *_Bucket {
	now := c.clock.Now()
	elapsed := now.Sub(c.lastAccess)
	count := int(elapsed / c.window)
	if count > 0 {
		c.lastAccess = now
		if count >= len(c.buckets) {
			c.resetAllBuckets()
		} else {
			// Reset the buckets between now and number of buckets ago. If
			// that is more that the existing buckets, reset all.
			for i := 0; i < count; i++ {
				c.offset++
				if c.offset >= len(c.buckets) {
					c.offset = 0
				}
				c.resetBucket(c.buckets[c.offset])
			}
		}
	}

	return c.buckets[c.offset]
}

func (c *_Counter) resetAllBuckets() {
	c.total.success = 0
	c.total.failure = 0
	c.offset = 0
	for _, b := range c.buckets {
		b.Reset()
	}
}

func (c *_Counter) resetBucket(b *_Bucket) {
	t := &c.total
	t.success -= b.success
	t.failure -= b.failure

	b.Reset()
}
