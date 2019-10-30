package cli

type NewOptions struct {
	Comma    rune    // 解析命令行使用
	AutoExec bool    // 自动使用os.Args执行Exec,默认true
	Cmds     []*Cmd  // 命令
	Flags    []*Flag //
}

type NewOption func(o *NewOptions)

func WithComma(comma rune) NewOption {
	return func(o *NewOptions) {
		o.Comma = comma
	}
}

func DisableAutoExec() NewOption {
	return func(o *NewOptions) {
		o.AutoExec = false
	}
}

func WithCmds(cmd ...*Cmd) NewOption {
	return func(o *NewOptions) {
		o.Cmds = cmd
	}
}

func WithFlags(flags ...*Flag) NewOption {
	return func(o *NewOptions) {
		o.Flags = flags
	}
}
