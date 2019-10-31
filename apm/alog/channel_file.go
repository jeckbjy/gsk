package alog

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// 文件输出,rotate strategy
type FileChannel struct {
	BaseChannel
	path string
	err  error
	file *os.File
}

func (c *FileChannel) Name() string {
	return "file"
}

func (c *FileChannel) SetConfig(key string, value string) error {
	return nil
}

func (c *FileChannel) Write(l *Logger, msg *Entry) {
	if c.Open() {
		text := msg.Format(c.formatter)
		if text != nil {
			_, _ = c.file.Write(text)
		}
	}
}

func (c *FileChannel) Open() bool {
	if c.file == nil && c.err == nil {
		if c.path == "" {
			c.path = fmt.Sprintf("%s.log", filepath.Base(os.Args[0]))
		}

		dir := path.Dir(c.path)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			c.err = err
			return false
		}
		c.file, err = os.OpenFile(c.path, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			c.err = err
			c.file = nil
		}
	}

	return c.file != nil
}
