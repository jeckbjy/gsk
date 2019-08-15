package frame

import (
	"errors"

	"github.com/jeckbjy/gsk/util/buffer"
)

// Default default frame codec
var Default IFrame

var ErrOverflow = errors.New("frame overflow")
var ErrIncomplete = errors.New("frame data incomplete")

var gFrameMap = make(map[string]IFrame)

func Add(f IFrame) {
	gFrameMap[f.Name()] = f
	Default = f
}

func Get(name string) IFrame {
	return gFrameMap[name]
}

// IFrame 用于粘包处理
type IFrame interface {
	Name() string
	Encode(b *buffer.Buffer) error
	Decode(b *buffer.Buffer) (*buffer.Buffer, error)
}
