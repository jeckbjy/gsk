package cli

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jeckbjy/gsk/util/strx"
)

// TODO:没有实现help格式化输出功能
type _Engine struct {
	root []Command // 以树形形式组织,只有叶节点才可以执行
	list []Command // 所有可执行节点
}

func (e *_Engine) List() []Command {
	return nil
}

func (e *_Engine) Tree() []Command {
	return e.root
}

func (e *_Engine) Add(action Action, opts ...Option) error {
	// 反射信息
	o := Options{}
	for _, fn := range opts {
		fn(&o)
	}

	if o.Name == "" {
		// example:RankTopCommand or RankTopCmd ====> rank_top
		t := reflect.TypeOf(action)
		name := t.Elem().Name()
		if strings.HasSuffix(name, "Cmd") {
			name = name[:len(name)-3]
		} else if strings.HasSuffix(name, "Command") {
			name = name[:len(name)-len("Command")]
		}
		o.Name = strx.ToSnake(name)
	}

	o.Name = strings.TrimLeft(o.Name, "/")
	full := trimSpaceAll(strings.Split(o.Name, "/"))
	if len(full) == 0 {
		return errors.New("invalid command name")
	}

	cmd := &_Command{name: full[len(full)-1], full: o.Name, group: o.Group, action: action, meta: o.Meta}
	if err := cmd.Parse(action); err != nil {
		return err
	}

	if len(full) == 1 {
		e.root = append(e.root, cmd)
		e.list = append(e.list, cmd)
	} else {
		parent, err := e.findParentCommand(full)
		if err != nil {
			return err
		}
		parent.Add(cmd)
		e.list = append(e.list, cmd)
	}

	return nil
}

func (e *_Engine) findParentCommand(path []string) (Command, error) {
	last := findCommand(e.root, path[0])
	if last == nil {
		last = &_Command{name: path[0], full: path[0]}
		e.root = append(e.root, last)
	} else if last.CanRun() {
		return nil, ErrNameConflict
	}

	if len(path) == 2 {
		return last, nil
	}

	for i := 1; i < len(path)-1; i++ {
		name := path[i]
		sub := findCommand(last.Subs(), name)
		if sub.CanRun() {
			return nil, ErrNameConflict
		}
		if sub == nil {
			sub = &_Command{name: name, full: fmt.Sprintf("%s/%s", last.Name(), name)}
			last.Add(sub)
		}
		last = sub
	}

	return last, nil
}

func (e *_Engine) Exec(args []string, metas map[string]string) (interface{}, error) {
	parser := Parser{}
	if err := parser.Parse(args); err != nil {
		return nil, err
	}

	if len(parser.params) == 0 {
		return nil, ErrNoCommand
	}

	// find and run command
	var cmd Command
	commands := e.root
	params := parser.params
	for i := 0; i < len(params); i++ {
		sub := findCommand(commands, params[i])
		if sub == nil {
			return nil, ErrNotFoundCommand
		}

		if sub.CanRun() {
			// remove cmd names
			cmd = sub
			params = params[i+1:]
			break
		}

		commands = sub.Subs()
	}

	if cmd == nil {
		return nil, ErrNotFoundCommand
	}

	ctx := &_Context{params: params, flags: parser.options, metas: metas}
	err := cmd.Run(ctx)

	return ctx.Result(), err
}

func findCommand(commands []Command, name string) Command {
	for _, cmd := range commands {
		if cmd.Name() == name {
			return cmd
		}
	}

	return nil
}
