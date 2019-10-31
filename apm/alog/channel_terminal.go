package alog

import (
	"io"
	"os"
)

func NewTerminal() Channel {
	return &TerminalChannel{writer: os.Stdout}
}

// 控制台输出
type TerminalChannel struct {
	BaseChannel
	writer io.Writer
}

func (c *TerminalChannel) Name() string {
	return "terminal"
}

func (c *TerminalChannel) SetProperty(key string, value string) error {
	switch key {
	case "stderr":
		c.writer = os.Stderr
	default:
		return c.BaseChannel.SetProperty(key, value)
	}
	return nil
}

func (c *TerminalChannel) Write(msg *Entry) {
	text := msg.Format(c.formatter)
	if text != nil {
		_, _ = c.writer.Write(text)
	}
}
