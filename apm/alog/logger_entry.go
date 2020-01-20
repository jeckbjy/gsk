package alog

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Entry struct {
	*runtime.Frame
	logger    *Logger
	formatter Formatter
	Time      time.Time
	Level     Level
	Text      string
	Fields    map[string]string
	filename  string
	next      *Entry
	outputs   map[int][]byte
}

// 重置数据,复用
func (e *Entry) Reset() {
	e.filename = ""
	e.next = nil
	e.outputs = make(map[int][]byte)
}

// 返回文件名,不包含名字
func (e *Entry) FileName() string {
	if e.filename == "" {
		e.filename = filepath.Base(e.File)
	}

	return e.filename
}

func (e *Entry) Source(short bool) string {
	if short {
		return fmt.Sprintf("%s:%d", e.FileName(), e.Line)
	} else {
		return fmt.Sprintf("%s:%d", e.File, e.Line)
	}
}

// 调用方法名
func (e *Entry) FuncName() string {
	idx := strings.LastIndexByte(e.Function, '.')
	if idx > 0 {
		return e.Function[idx+1:]
	}

	return ""
}

func (e *Entry) FuncLine() string {
	return fmt.Sprintf("%s:%d", e.FuncName(), e.Line)
}

func (e *Entry) GetField(key string) string {
	if e.Fields != nil {
		if v, ok := e.Fields[key]; ok {
			return v
		}
	}

	return e.logger.GetField(key)
}

func (e *Entry) Format(f Formatter) []byte {
	if f == nil {
		f = e.formatter
	}
	if data, ok := e.outputs[f.ID()]; ok {
		return data
	}

	data, err := f.Format(e)
	if err != nil {
		return nil
	}
	if e.outputs == nil {
		e.outputs = make(map[int][]byte)
	}
	e.outputs[f.ID()] = data
	return data
}
