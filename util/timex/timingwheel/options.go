package timingwheel

type Option func(o *Options)
type Options struct {
	interval int64 // 间隔,精度
}

// 时间间隔
func Interval(t int64) Option {
	return func(o *Options) {
		o.interval = t
	}
}
