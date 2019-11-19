package bi

import (
	"errors"
)

var (
	ErrNotInit    = errors.New("not init")
	ErrHasInit    = errors.New("has init")
	ErrTooManyMsg = errors.New("too many msg")
	ErrBadOption  = errors.New("bad option")
	ErrBadEvent   = errors.New("bad event")
)

// global reporter
var gReporter Reporter

// 初始化数据
func Init(opts *Options) error {
	return gReporter.Init(opts)
}

func Send(event string, params M) error {
	return gReporter.Send(event, params)
}

// 通过反射类发送
func Sendx(s interface{}) error {
	event, params, err := Reflect(s)
	if err != nil {
		return err
	}

	return gReporter.Send(event, params)
}
