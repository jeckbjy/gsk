package cli

import (
	"context"
	"errors"
)

func newContext(ctx context.Context, engine *Engine, cmd *Cmd, args []string, flags map[string][]string) *Context {
	return &Context{Context: ctx, engine: engine, cmd: cmd, args: args, flags: flags}
}

type Context struct {
	context.Context
	engine *Engine
	cmd    *Cmd                // cmd
	args   []string            // cmd参数
	flags  map[string][]string // 可选数据,可以为空
}

// args
func (c *Context) NArg() int {
	return len(c.args)
}

func (c *Context) Arg(i int) string {
	return c.args[i]
}

func (c *Context) BindArg(i int, data interface{}) error {
	if i < len(c.args) {
		return bindValue(c.args[i], data)
	}

	return errors.New("overflow")
}

// flags
func (c *Context) NFlag() int {
	return len(c.flags)
}

func (c *Context) Flag(key string) []string {
	return c.flags[key]
}

// 通过key绑定flag
func (c *Context) BindFlag(key string, data interface{}) {

}

// 绑定args和flag,通过类型判断
func (c *Context) Bind() {

}
