package store

type Options struct {
	Prefix   bool  // Delete,Watch use
	KeyOnly  bool  // Get,List
	Revision int64 // Get,List
}

func (o *Options) Build(opts ...Option) {
	for _, fn := range opts {
		fn(o)
	}
}

type Option func(*Options)

func Prefix() Option {
	return func(o *Options) {
		o.Prefix = true
	}
}

func KeyOnly() Option {
	return func(o *Options) {
		o.KeyOnly = true
	}
}

func Revision(r int64) Option {
	return func(o *Options) {
		o.Revision = r
	}
}
