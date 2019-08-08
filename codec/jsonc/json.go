package jsonc

import (
	"encoding/json"
	"github.com/jeckbjy/micro/codec"
	"github.com/jeckbjy/micro/util/buffer"
)

func init() {
	codec.Default = New()
}

// New create json codec
func New() codec.ICodec {
	return &Codec{}
}

type Codec struct {
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
