package errorx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jeckbjy/gsk/util/idgen/sid"
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
	ErrInvalidConfig = errors.New("invalid config")
)

// 扩展error接口,方便外部使用
// ID:全局唯一ID,用于追踪查询
// Code:与http编码保持一致
// Status: 包含两部分,http默认信息以及自定义信息,逗号分隔
// Unwrap: 返回原始错误信息,比如DB error
// Debug:  会以json形式返回更加详细的错误信息
// Error:  仅返回错误编号和ID,保证客户端不会窃取到服务器调试信息
type Error interface {
	error
	ID() string
	Code() int
	Status() string
	Unwrap() error
	Debug() string
}

// New 创建Error
func New(err error, code int, format string, args ...interface{}) Error {
	return NewWithSkip(3, err, code, format, args...)
}

// 区别于New,可以外部指定skip
func NewWithSkip(skip int, err error, code int, format string, args ...interface{}) Error {
	id, _ := sid.Generate()
	builder := strings.Builder{}
	if code <= http.StatusNetworkAuthenticationRequired {
		status := http.StatusText(code)
		if status != "" {
			builder.WriteString("[")
			builder.WriteString(status)
			builder.WriteString("]")
		}
	}
	detail := fmt.Sprintf(format, args...)
	if detail != "" {
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(detail)
	}

	return &xerror{
		XID:     id,
		XCode:   code,
		XStatus: builder.String(),
		XCaller: getCaller(skip),
		err:     err,
	}
}

type xerror struct {
	XID     string `json:"id"`
	XCode   int    `json:"code"`
	XStatus string `json:"status"`
	XErr    string `json:"err,omitempty"`
	XCaller string `json:"caller"`
	err     error
}

func (e *xerror) ID() string {
	return e.XID
}

func (e *xerror) Code() int {
	return e.XCode
}

func (e *xerror) Status() string {
	return e.XStatus
}

func (e *xerror) Unwrap() error {
	return e.err
}

func (e *xerror) Error() string {
	return fmt.Sprintf("%d %s", e.XCode, e.XID)
}

func (e *xerror) Debug() string {
	if e.XErr == "" && e.err != nil {
		e.XErr = e.err.Error()
	}
	b, _ := json.Marshal(e)
	return string(b)
}

func getCaller(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	f := runtime.FuncForPC(pc)
	file, line := f.FileLine(pc)
	name := f.Name()
	if idx := strings.LastIndexByte(name, '.'); idx != -1 {
		name = name[idx+1:]
	}
	return fmt.Sprintf("%s:%d[%s]", filepath.Base(file), line, name)
}
