package len32

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/jeckbjy/micro/frame"
	"github.com/jeckbjy/micro/util/buffer"
)

func init() {
	frame.Default = New()
}

func New() frame.IFrame {
	return &Frame{}
}

type Frame struct {
	limit int
}

func (*Frame) Encode(b *buffer.Buffer) error {
	if b.Len() > math.MaxUint32 {
		return frame.ErrOverflow
	}

	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(b.Len()))
	b.Prepend(data)
	return nil
}

func (f *Frame) Decode(b *buffer.Buffer) (*buffer.Buffer, error) {
	data := [4]byte{}
	if n, _ := b.Read(data[:]); n != 4 {
		return nil, frame.ErrIncomplete
	}

	size := binary.LittleEndian.Uint32(data[:])
	if b.Len()-b.Pos() < int(size) {
		return nil, frame.ErrIncomplete
	}

	if f.limit > 0 && int(size) > f.limit {
		return nil, frame.ErrOverflow
	}

	// 去除长度
	b.Discard()
	_, _ = b.Seek(int64(size), io.SeekStart)
	d := b.Split()
	return d, nil
}
