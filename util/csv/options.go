package csv

type Option func(o *Options)
type Options struct {
	Comma        rune     // 分隔符,默认逗号
	IgnorePrompt rune     // 忽略该行提示符,默认为#,0表示不忽略
	IgnoreLines  int      // 忽略行数,默认三行,第一行名字,第二行类型,第三行注释
	NameLine     int      // 名字所在行号,默认为0,-1表示没有头信息
	NameHead     []string // 如果数据不提供表头信息,则需要手动提供每一列字段名
}

func (o *Options) Init(opts ...Option) {
	o.Comma = ','
	o.IgnorePrompt = '#'
	o.IgnoreLines = 3
	o.NameLine = 0
	for _, fn := range opts {
		fn(o)
	}
}
