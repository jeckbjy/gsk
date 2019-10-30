package codec

import "github.com/jeckbjy/gsk/util/buffer"

var gCodecMap = make(map[string]Codec)

func Add(c Codec) {
	gCodecMap[c.Name()] = c
}

func Get(name string) Codec {
	return gCodecMap[name]
}

// Codec 消息编解码
type Codec interface {
	Name() string
	Encode(b *buffer.Buffer, msg interface{}) error
	Decode(b *buffer.Buffer, msg interface{}) error
}
