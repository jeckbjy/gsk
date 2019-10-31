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
	formatter Formatter
	Time      time.Time
	Level     Level
	Text      string
	Fields    map[string]string
	filename  string
	next      *Entry
	outputs   map[int][]byte
}

func (e *Entry) FileName() string {
	if e.filename == "" {
		e.filename = filepath.Base(e.File)
	}

	return e.filename
}

func (e *Entry) FileLine() string {
	return fmt.Sprintf("%s:%d", e.FileName(), e.Line)
}

func (e *Entry) Method() string {
	idx := strings.LastIndexByte(e.Function, '.')
	if idx > 0 {
		return e.Function[idx+1:]
	}

	return ""
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
