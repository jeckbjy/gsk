package alog

import (
	"fmt"
	"strings"
)

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelOff
	LevelInherit // 从logger继承
)

type Level uint8

func (l Level) String() string {
	switch l {
	case LevelFatal:
		return "Fatal"
	case LevelError:
		return "Error"
	case LevelWarn:
		return "Warn"
	case LevelInfo:
		return "Info"
	case LevelDebug:
		return "Debug"
	case LevelTrace:
		return "Trace"
	case LevelOff:
		return "Off"
	default:
		return "Unknown"
	}
}

func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(value) {
	case "fatal":
		return LevelFatal, nil
	case "error":
		return LevelError, nil
	case "warn":
		return LevelWarn, nil
	case "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	case "trace":
		return LevelTrace, nil
	case "off":
		return LevelOff, nil
	default:
		return LevelOff, fmt.Errorf("unknown level,%+v", value)
	}
}
