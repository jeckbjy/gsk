package jsonc

import (
	"encoding/json"

	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

// New create json codec
func New() codec.Codec {
	return &Codec{}
}

type Codec struct {
}

func (*Codec) Type() int {
	return codec.Json
}

func (*Codec) Name() string {
	return "json"
}

func (*Codec) Encode(b *buffer.Buffer, msg interface{}) error {
	e := json.NewEncoder(b)
	return e.Encode(msg)
}

func (*Codec) Decode(b *buffer.Buffer, msg interface{}) error {
	d := json.NewDecoder(b)
	return d.Decode(msg)
}
