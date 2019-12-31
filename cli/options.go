package cli

type Option func(o *Options)
type Options struct {
	Name  string            // command名字,可以是xx/xx/xxx形式
	Group string            // command分组,可用于help格式化输出
	Meta  map[string]string // 自定义数据用于扩展
}

func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

func Group(g string) Option {
	return func(o *Options) {
		o.Group = g
	}
}

func Meta(meta map[string]string) Option {
	return func(o *Options) {
		o.Meta = meta
	}
}
