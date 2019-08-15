// protobuf codec
package protoc

import (
	"errors"
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/codec/protoc/proto"
	"github.com/jeckbjy/gsk/util/buffer"
)

func init() {
	codec.Add(New())
}

var ErrNotMessage = errors.New("not pb message")

func New() codec.ICodec {
	return &Codec{}
}

type Codec struct {
}

func (*Codec) Name() string {
	return "proto"
}

func (*Codec) Encode(b *buffer.Buffer, msg interface{}) error {
	if m, ok := msg.(proto.Message); ok {
		data, err := proto.Marshal(m)
		if err != nil {
			return err
		}
		b.Append(data)
		return nil
	}

	return ErrNotMessage
}

func (*Codec) Decode(b *buffer.Buffer, msg interface{}) error {
	if pb, ok := msg.(proto.Message); ok {
		data := b.Bytes()
		return proto.Unmarshal(data, pb)
	}

	return ErrNotMessage
}
