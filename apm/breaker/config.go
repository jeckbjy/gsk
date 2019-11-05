package breaker

import "time"

const (
	defaultRequest = 20
	defaultRatio   = 0.5
	defaultSleep   = time.Millisecond * 5000
	defaultWindow  = time.Millisecond * 10000
	defaultBucket  = 10
)

type Config struct {
	Disabled         bool          // 是否禁用
	RequestThreshold int           // 超过此值才会触发熔断校验,区间段为Window总和
	Ratio            float64       // 短路百分比,默认50%
	Sleep            time.Duration // 休眠时间,即open状态经过sleep时间后,进入HalfOpen状态
	Window           time.Duration // 窗口时间
	Bucket           int           // 窗口桶个数
	clock            _Clock        // 用于测试
}

func (c *Config) fix() {
	if c.RequestThreshold <= 0 {
		c.RequestThreshold = defaultRequest
	}

	if c.Ratio <= 0 {
		c.Ratio = defaultRatio
	}

	if c.Sleep <= 0 {
		c.Sleep = defaultSleep
	}

	if c.Window <= 0 {
		c.Window = defaultWindow
	}

	if c.Bucket <= 0 {
		c.Bucket = defaultBucket
	}

	if c.clock == nil {
		c.clock = gTimeClock
	}
}
