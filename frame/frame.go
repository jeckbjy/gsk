package frame

import (
	"errors"
	"github.com/jeckbjy/micro/util/buffer"
)

// Default default frame codec
var Default IFrame

var ErrOverflow = errors.New("frame overflow")
var ErrIncomplete = errors.New("frame data incomplete")

// IFrame 用于粘包处理
type IFrame interface {
	Encode(b *buffer.Buffer) error
	Decode(b *buffer.Buffer) (*buffer.Buffer, error)
}
