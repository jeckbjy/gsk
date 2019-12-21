package codec

import "github.com/jeckbjy/gsk/util/buffer"

var gCodecMap = make(map[string]Codec)
var gTypeList = make([]Codec, Gob+1)

func Add(c Codec) {
	gCodecMap[c.Name()] = c
	if c.Type() > len(gTypeList) {
		count := len(gTypeList) - c.Type() + 1
		gTypeList = append(gTypeList, make([]Codec, count)...)
	}

	gTypeList[c.Type()] = c
}

func GetByName(name string) Codec {
	return gCodecMap[name]
}

func Get(t int) Codec {
	if t < len(gTypeList) {
		return gTypeList[t]
	}

	return nil
}

func SetDefault(c Codec) {
	gTypeList[Default] = c
}

// 枚举定义常见的消息编码格式
const (
	Default = iota
	Json
	Proto
	Xml
	Gob
)

// Codec 消息编解码
type Codec interface {
	Type() int
	Name() string
	Encode(b *buffer.Buffer, msg interface{}) error
	Decode(b *buffer.Buffer, msg interface{}) error
}
