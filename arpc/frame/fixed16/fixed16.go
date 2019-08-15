package fixed16

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/jeckbjy/gsk/frame"
	"github.com/jeckbjy/gsk/util/buffer"
)

func init() {
	frame.Add(New())
}

func New() frame.IFrame {
	return &Frame{}
}

type Frame struct {
	limit int
}

func (f *Frame) SetLimit(limit int) {
	f.limit = limit
}

func (f *Frame) Name() string {
	return "fixed16"
}

func (f *Frame) Encode(b *buffer.Buffer) error {
	if b.Len() > math.MaxUint16 {
		return frame.ErrOverflow
	}

	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(b.Len()))
	b.Prepend(data)
	return nil
}

func (f *Frame) Decode(b *buffer.Buffer) (*buffer.Buffer, error) {
	data := [2]byte{}
	if n, _ := b.Read(data[:]); n != 2 {
		return nil, frame.ErrIncomplete
	}

	size := binary.LittleEndian.Uint16(data[:])
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
