package gobc

import (
	"encoding/gob"

	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

const Name = "gob"

// New create gob codec
func New() codec.Codec {
	return &Codec{}
}

type Codec struct {
}

func (*Codec) Type() int {
	return codec.Gob
}

func (*Codec) Name() string {
	return Name
}

func (*Codec) Encode(b *buffer.Buffer, msg interface{}) error {
	e := gob.NewEncoder(b)
	return e.Encode(msg)
}

func (*Codec) Decode(b *buffer.Buffer, msg interface{}) error {
	d := gob.NewDecoder(b)
	return d.Decode(msg)
}
