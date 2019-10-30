package cli

type Flag struct {
	Name     string // --help
	Shortcut string // -h
	Default  string // default value
	Usage    string // describe
	Param    bool   // 是否含有参数
	Multiple bool   // 是否可以有多个参数
}
