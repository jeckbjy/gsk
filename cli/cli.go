package cli

import (
	"context"
	"fmt"
	"os"
)

func New(opts ...NewOption) *Engine {
	e := &Engine{root: &Cmd{}}
	o := NewOptions{}
	o.Comma = ' '
	o.AutoExec = true
	for _, fn := range opts {
		fn(&o)
	}
	e.comma = o.Comma
	e.autoExec = o.AutoExec
	if len(o.Cmds) > 0 {

	}

	if len(o.Flags) > 0 {

	}

	return e
}

type Engine struct {
	comma    rune
	autoExec bool
	root     *Cmd
}

// 解析并执行command,param要求必须是string,或者[]string
// 如果param是string,则表示原生的一行数据,需要使用comma分隔解析成[]string
// 如果param是[]string,则表示不需额外解析
//
// 解析规则:类似unix下命令行参数解析,-作为先行引导符,表示flag
// cmd解析:
// 每个独立的token则是一个cmd
// flag解析:
// 1:-表示shortcut,可以多个合并,例如,-h 表示help, -czvx，表示-c -z -v -x,如果带参数,则只会设置给最后一个
// 2:--表示全称，例如--help
// 3:后边可以紧跟一个参数,可以使用=连载一起写,也可以空格分隔
// 4:flag可以重复,相同的则合并成1个处理
//
// 使用限制要求:需要将command命令和参数写在前边,flag写在后边,否则将会把cmd参数当成flag的参数
func (e *Engine) Exec(ctx context.Context, param interface{}) error {
	var argv []string
	switch param.(type) {
	case string:
		if fields, err := parseCommandLine(param.(string), e.comma); err != nil {
			return err
		} else {
			argv = fields
		}
	case []string:
		argv = param.([]string)
	default:
		return fmt.Errorf("bad param type,%+v", param)
	}

	args, err := parseArgs(argv)
	if err != nil {
		return err
	}

	// run command
	cmd, err := e.root.find(args.Params)
	if err != nil {
		return err
	}

	funcCtx := newContext(ctx, e, cmd, args.Params[cmd.Depth:], args.Options)
	return cmd.Func(funcCtx)
}

// 初始化,如果是控制台则会自动执行os.Args
func (e *Engine) Run() error {
	return e.Exec(context.Background(), os.Args[1:])
}

// 注册Command,会使用反射解析Command
func (e *Engine) Register() {

}
