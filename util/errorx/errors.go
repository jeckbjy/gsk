package errorx

import (
	"encoding/json"
	"errors"
)

var (
	ErrNotSupport    = errors.New("not support")
	ErrNotFound      = errors.New("not found")
	ErrNotAvailable  = errors.New("not available")
	ErrNotReady      = errors.New("not ready")
	ErrHasStopped    = errors.New("has stopped")
	ErrHasRegistered = errors.New("has registered")
	ErrInvalidId     = errors.New("invalid id")
	ErrInvalidParam  = errors.New("invalid param")
	ErrNoConfig      = errors.New("no config")
)

// 扩展error接口,方便外部使用
// 错误信息需要细分,比如DB错误,系统错误,IO错误,逻辑错误
// 上层逻辑可以根据错误的类型做不同的处理,
// 比如严重的系统需要上报,客户端业务错误可以单纯记录日志,并通知前端
// 错误可以有一个全局唯一ID,方便追踪查询
// 错误信息可以记录一些当时的环境变量,方便复现与调试
// Unwrap可以获取原始错误信息,比如DB操作错误
type Error interface {
	error
	Category() int
	ID() string
	Code() int
	Status() string
	Detail() string
	Unwrap() error
}

func New() Error {
	return &_UniformError{}
}

type _UniformError struct {
	category int
	id       string
	code     int
	status   string
	detail   string
	raw      error
}

func (e *_UniformError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e *_UniformError) Category() int {
	return e.category
}

func (e *_UniformError) ID() string {
	return e.id
}

func (e *_UniformError) Code() int {
	return e.code
}

func (e *_UniformError) Status() string {
	return e.status
}

func (e *_UniformError) Detail() string {
	return e.detail
}

func (e *_UniformError) Unwrap() error {
	return e.raw
}
