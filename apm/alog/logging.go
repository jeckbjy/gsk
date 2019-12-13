package alog

var std = New()

func WithFields(fields map[string]string) *Builder {
	return std.WithFields(fields)
}

func Trace(args ...interface{}) {
	std.Log(LevelTrace, args...)
}

func Tracef(format string, args ...interface{}) {
	std.Logf(LevelTrace, format, args...)
}

func Debug(args ...interface{}) {
	std.Log(LevelDebug, args...)
}

func Debugf(format string, args ...interface{}) {
	std.Logf(LevelDebug, format, args...)
}

func Info(args ...interface{}) {
	std.Log(LevelInfo, args...)
}

func Infof(format string, args ...interface{}) {
	std.Logf(LevelInfo, format, args...)
}

func Warn(args ...interface{}) {
	std.Log(LevelWarn, args...)
}

func Warnf(format string, args ...interface{}) {
	std.Logf(LevelWarn, format, args...)
}

func Error(args ...interface{}) {
	std.Log(LevelError, args...)
}

func Errorf(format string, args ...interface{}) {
	std.Logf(LevelError, format, args...)
}

func Fatal(args ...interface{}) {
	std.Log(LevelFatal, args...)
}

func Fatalf(format string, args ...interface{}) {
	std.Logf(LevelFatal, format, args...)
}
