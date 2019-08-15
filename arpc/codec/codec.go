package codec

import "github.com/jeckbjy/gsk/util/buffer"

var Default ICodec
var gCodecMap = make(map[string]ICodec)

func Add(c ICodec) {
	gCodecMap[c.Name()] = c
	Default = c
}

func Get(name string) ICodec {
	return gCodecMap[name]
}

// ICodec 消息编解码
type ICodec interface {
	Name() string
	Encode(b *buffer.Buffer, msg interface{}) error
	Decode(b *buffer.Buffer, msg interface{}) error
}
