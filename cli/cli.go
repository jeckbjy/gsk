package cli

import "errors"

var (
	ErrNoCommand       = errors.New("no command")
	ErrNotFoundCommand = errors.New("not found command")
	ErrNotSupport      = errors.New("not support")
	ErrInvalidTag      = errors.New("invalid tag")
	ErrNameConflict    = errors.New("name conflict")
)

type Context interface {
	NArg() int
	Arg(index int) string
	NFlag() int
	Flag(key string) []string
	Get(key string) string
	// 返回结果
	Result() interface{}
	JSON(value interface{}) error
	XML(value interface{}) error
	Text(format string, values ...interface{}) error
	Any(value interface{}) error
}

// 最终需要执行的命令
type Action interface {
	Run(ctx Context) error
}

// Command 以树形结构组织,只有叶节点可以可以执行
type Command interface {
	Name() string
	Full() string
	Group() string
	Meta() map[string]string
	Subs() []Command
	Add(sub Command)
	CanRun() bool
	Run(ctx Context) error
}

// Engine 用于管理所有Command信息
// List只返回所有可执行的Command
// Tree会以树状的形式组织Command,只有叶节点可以执行
type Engine interface {
	List() []Command
	Tree() []Command
	Add(action Action, opts ...Option) error
	Exec(args []string, metas map[string]string) (interface{}, error)
}

// New 创建Engine
func New() Engine {
	return &_Engine{}
}
