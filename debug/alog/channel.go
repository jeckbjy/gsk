package alog

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
	Write(msg *Entry)
}
