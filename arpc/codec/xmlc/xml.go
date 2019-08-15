package xmlc

import (
	"encoding/xml"
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

func init() {
	codec.Add(New())
}

// New create xml codec
func New() codec.ICodec {
	return &Codec{}
}

type Codec struct {
}

func (*Codec) Name() string {
	return "xml"
}

func (*Codec) Encode(b *buffer.Buffer, msg interface{}) error {
	e := xml.NewEncoder(b)
	return e.Encode(msg)
}

func (*Codec) Decode(b *buffer.Buffer, msg interface{}) error {
	d := xml.NewDecoder(b)
	return d.Decode(msg)
}
