package ebus

type Option func(*Options)

type Options struct {
	topic    string
	once     bool
	priority int
}

func Topic(t string) Option {
	return func(o *Options) {
		o.topic = t
	}
}

func Once() Option {
	return func(o *Options) {
		o.once = true
	}
}

func Priority(lv int) Option {
	return func(o *Options) {
		o.priority = lv
	}
}
