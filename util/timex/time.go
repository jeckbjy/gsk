package timex

import (
	"time"

	"github.com/jeckbjy/gsk/util/timex/timingwheel"
)

type Timer = timingwheel.Timer

// Now 返回毫秒值
func Now() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// NowUnix 返回秒
func NowUnix() int64 {
	return time.Now().Unix()
}

// NewTimer 创建定时器
func NewTimer(expired time.Duration, task func()) *Timer {
	return timingwheel.NewTimer(int64(expired/time.Millisecond), task)
}

// 通过过期时间创建timer
func NewDelayTimer(delay time.Duration, task func()) *Timer {
	expired := time.Now().Add(delay).UnixNano() / int64(time.Millisecond)
	return timingwheel.NewTimer(expired, task)
}
