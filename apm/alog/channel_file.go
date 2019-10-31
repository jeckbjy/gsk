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

func (c *FileChannel) SetProperty(key string, value string) error {
	switch key {
	case "path":
		if c.path != value {
			c.path = value
			_ = c.Close()
		}
	default:
		return c.BaseChannel.SetProperty(key, value)
	}
	return nil
}

func (c *FileChannel) Open() error {
	if c.file == nil && c.err == nil {
		if c.path == "" {
			c.path = fmt.Sprintf("%s.log", filepath.Base(os.Args[0]))
		}

		dir := path.Dir(c.path)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			c.err = err
			return err
		}
		c.file, err = os.OpenFile(c.path, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			c.err = err
			c.file = nil
		}
	}

	if c.file == nil {
		return ErrNotReady
	}

	return nil
}

func (c *FileChannel) Close() error {
	if c.file != nil {
		err := c.file.Close()
		c.file = nil
		return err
	}

	return nil
}

func (c *FileChannel) Write(msg *Entry) {
	if c.Open() == nil {
		text := msg.Format(c.formatter)
		if text != nil {
			_, _ = c.file.Write(text)
		}
	}
}
