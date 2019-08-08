package varint

import (
	"encoding/binary"
	"io"

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
	d := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(d, uint64(b.Len()))
	b.Prepend(d[:n])
	return nil
}

func (f *Frame) Decode(b *buffer.Buffer) (*buffer.Buffer, error) {
	size, err := binary.ReadUvarint(b)
	if err != nil {
		return nil, err
	}

	if f.limit > 0 && int(size) > f.limit {
		return nil, frame.ErrOverflow
	}

	if b.Len()-b.Pos() < int(size) {
		return nil, frame.ErrIncomplete
	}

	// 去除长度
	b.Discard()
	_, _ = b.Seek(int64(size), io.SeekStart)
	d := b.Split()
	return d, nil
}
