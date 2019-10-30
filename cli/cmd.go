package cli

import "errors"

var (
	errNotFoundCmd   = errors.New("not found cmd")
	errNoLeafNodeCmd = errors.New("no leaf node cmd")
	errDuplicateFlag = errors.New("duplicate flag name")
)

type Action func(ctx *Context) error
type Cmd struct {
	Depth   int
	Name    string
	Flags   []*Flag
	Subs    []*Cmd
	Func    Action
	parent  *Cmd             // 父节点
	flagMap map[string]*Flag // 所有flag,包括继承父节点的,同时包括shortcut,只有叶节点才需要
}

func (c *Cmd) Parent() *Cmd {
	return c.parent
}

// 递归查找sub commond
func (c *Cmd) find(keys []string) (*Cmd, error) {
	for _, v := range c.Subs {
		if v.Name == keys[0] {
			if len(v.Subs) == 0 {
				// 找到第一个没有sub的,则表示已经搜索到叶节点
				return v, nil
			} else if len(keys) == 1 {
				// 找到一个,但是不是叶节点
				return v, errNoLeafNodeCmd
			} else {
				// 继续查找
				return c.find(keys[1:])
			}
		}
	}

	return nil, errNotFoundCmd
}

// 构建parent
func (c *Cmd) build() error {
	if len(c.Subs) > 0 {
		for _, v := range c.Subs {
			v.parent = c
			v.Depth = c.Depth + 1
		}
	} else {
		// leaf node, build flag map, and inherit from parent
		c.flagMap = make(map[string]*Flag)
		if err := c.buildFlagMap(c.flagMap); err != nil {
			return err
		}

		// inherit from parent
		node := c.parent
		for node != nil {
			if err := node.buildFlagMap(c.flagMap); err != nil {
				return err
			}
			node = node.parent
		}
	}

	return nil
}

func (c *Cmd) buildFlagMap(flagMap map[string]*Flag) error {
	for _, f := range c.Flags {
		if f.Name != "" {
			if _, exists := flagMap[f.Name]; exists {
				return errDuplicateFlag
			}
			flagMap[f.Name] = f
		}

		if f.Shortcut != "" {
			if _, exists := flagMap[f.Shortcut]; exists {
				return errDuplicateFlag
			}

			flagMap[f.Shortcut] = f
		}
	}

	return nil
}
