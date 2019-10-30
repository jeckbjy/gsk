package alog

var std = New()

func WithFields(fields map[string]string) *Builder {
	return std.WithFields(fields)
}

func Trace(args ...interface{}) {
	std.Log(LevelTrace, args...)
}
