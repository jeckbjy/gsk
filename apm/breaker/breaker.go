package breaker

import "errors"

var (
	ErrReject  = errors.New("reject")
	ErrTooMany = errors.New("too many request")
)

var gDefaultGroup = NewGroup(&Config{})

func Reload(cfg *Config) {
	gDefaultGroup.Reload(cfg)
}

func Get(key string) Breaker {
	return gDefaultGroup.Get(key)
}

func Exec(key string, run, fallback func() error) error {
	return gDefaultGroup.Exec(key, run, fallback)
}

// Breaker
//
// Breaker仅仅是一个状态机
// 使用Allow判断能否通过,使用MarkSuccess,MarkFailure反馈成功或失败结果
// Mark仅仅是MarkSuccess和MarkFailure的封装,方便外部操作
type Breaker interface {
	Allow() error
	Mark(err error)
	MarkSuccess()
	MarkFailure()
	String() string
}

// Breaker分组,共享一份配置
type Group interface {
	Get(key string) Breaker
	Reload(cfg *Config)
	Exec(key string, run, fallback func() error) error
}
