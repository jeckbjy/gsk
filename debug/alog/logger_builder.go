package alog

import "fmt"

type Builder struct {
	logger *Logger
	fields map[string]string
}

func (b *Builder) WithFields(fields map[string]string) *Builder {
	b.fields = fields
	return b
}

func (b *Builder) Log(lv Level, args ...interface{}) {
	b.logger.Write(lv, b.fields, 1, fmt.Sprint(args...))
}

func (b *Builder) Logf(lv Level, format string, args ...interface{}) {
	b.logger.Write(lv, b.fields, 1, fmt.Sprintf(format, args...))
}

func (b *Builder) Trace(args ...interface{}) {
	b.Log(LevelTrace, args...)
}

func (b *Builder) Debug(args ...interface{}) {
	b.Log(LevelDebug, args...)
}

func (b *Builder) Print(args ...interface{}) {
	b.Log(LevelPrint, args...)
}

func (b *Builder) Info(args ...interface{}) {
	b.Log(LevelInfo, args...)
}

func (b *Builder) Warn(args ...interface{}) {
	b.Log(LevelWarn, args...)
}

func (b *Builder) Error(args ...interface{}) {
	b.Log(LevelError, args...)
}

func (b *Builder) Fatal(args ...interface{}) {
	b.Log(LevelFatal, args...)
}

func (b *Builder) Tracef(format string, args ...interface{}) {
	b.Logf(LevelTrace, format, args...)
}

func (b *Builder) Debugf(format string, args ...interface{}) {
	b.Logf(LevelDebug, format, args...)
}

func (b *Builder) Printf(format string, args ...interface{}) {
	b.Logf(LevelPrint, format, args...)
}

func (b *Builder) Infof(format string, args ...interface{}) {
	b.Logf(LevelInfo, format, args...)
}

func (b *Builder) Warnf(format string, args ...interface{}) {
	b.Logf(LevelWarn, format, args...)
}

func (b *Builder) Errorf(format string, args ...interface{}) {
	b.Logf(LevelError, format, args...)
}

func (b *Builder) Fatalf(format string, args ...interface{}) {
	b.Logf(LevelFatal, format, args...)
}
