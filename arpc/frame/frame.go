package frame

import (
	"errors"
	"sync/atomic"

	"github.com/jeckbjy/gsk/util/buffer"
)

var ErrOverflow = errors.New("frame overflow")
var ErrIncomplete = errors.New("frame data incomplete")

var defaultFrame atomic.Value

func Default() Frame {
	return defaultFrame.Load().(Frame)
}

func SetDefault(r Frame) {
	defaultFrame.Store(r)
}

// Frame 用于粘包处理
type Frame interface {
	Name() string
	Encode(b *buffer.Buffer) error
	Decode(b *buffer.Buffer) (*buffer.Buffer, error)
}
