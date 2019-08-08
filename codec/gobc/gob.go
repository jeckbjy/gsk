package gobc

import (
	"encoding/gob"
	"github.com/jeckbjy/micro/codec"
	"github.com/jeckbjy/micro/util/buffer"
)

func init() {
	codec.Default = New()
}

// New create gob codec
func New() codec.ICodec {
	return &Codec{}
}

type Codec struct {
}

func (*Codec) Name() string {
	return "gob"
}

func (*Codec) Encode(b *buffer.Buffer, msg interface{}) error {
	e := gob.NewEncoder(b)
	return e.Encode(msg)
}

func (*Codec) Decode(b *buffer.Buffer, msg interface{}) error {
	d := gob.NewDecoder(b)
	return d.Decode(msg)
}
