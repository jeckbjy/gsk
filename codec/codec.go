package codec

import "github.com/jeckbjy/micro/util/buffer"

var Default ICodec

// ICodec 消息编解码
type ICodec interface {
	Name() string
	Encode(b *buffer.Buffer, msg interface{}) error
	Decode(b *buffer.Buffer, msg interface{}) error
}
