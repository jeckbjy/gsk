package file

type Options struct {
	Base string
}

type Option func(o *Options)

func Base(s string) Option {
	return func(o *Options) {
		o.Base = s
	}
}
