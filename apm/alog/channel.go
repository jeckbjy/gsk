package alog

import "fmt"

var (
	ErrNotReady   = fmt.Errorf("channel not ready")
	ErrNotSupport = fmt.Errorf("not support")
)

type Configurable interface {
	SetLevel(lv Level)
	GetLevel() Level
	SetFormatter(formatter Formatter)
	GetFormatter() Formatter
	SetProperty(key string, value string) error
	GetProperty(key string) string
}

type Channel interface {
	Configurable
	SetLogger(l *Logger)
	Name() string
	Open() error
	Close() error
	Write(msg *Entry)
}
