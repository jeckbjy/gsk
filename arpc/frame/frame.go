package frame

import (
	"errors"

	"github.com/jeckbjy/gsk/util/buffer"
)

var ErrOverflow = errors.New("frame overflow")
var ErrIncomplete = errors.New("frame data incomplete")

var gFrameMap = make(map[string]Frame)

func Add(f Frame) {
	gFrameMap[f.Name()] = f
}

func Get(name string) Frame {
	return gFrameMap[name]
}

// Frame 用于粘包处理
type Frame interface {
	Name() string
	Encode(b *buffer.Buffer) error
	Decode(b *buffer.Buffer) (*buffer.Buffer, error)
}
