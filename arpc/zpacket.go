package arpc

import (
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

// Packet 用于服务器间通信
// ID:表示唯一数值型ID,Name表示消息唯一名字,用于标识消息类型,二选一,通常使用唯一ID
// Service:RPC服务名,用于服务查询,Method:RPC方法名,用于通过函数名查询消息处理回调
// SeqID:用于异步RPC全局唯一ID,通常用于查询异步处理回调函数
// ID,Name,Method并不需要都存在,使用任意一个都可以查找到回调处理函数,但运行效率和注册方便程度上有所不同
// ID的方式节省空间，速度快，但需要额外的注册
// Name,Method的方式可以通过反射的方式，在注册消息处理的时候，自动解析出函数名和消息名
//
// 编码格式:消息长度+消息头(FLAG+DATA)+消息体
//   已知常见的消息头通过枚举定义,其他自定义消息头以string编码保存在map中
//
// 还有一些常用的消息头,比如TraceID,SpanID,Auth,Token
// 自定义数据需要 x-www-form-urlencoded？
type Packet interface {
	Reply() bool             // 标识是否是RPC应答请求
	Status() uint            // 类似HTTP的状态码,正整数
	ID() uint                // 消息ID,要求是正整数
	Name() string            // 消息名
	SeqID() string           // RPC全局SequenceID
	Method() string          // RPC方法名
	Service() string         // RPC服务名
	Head() map[string]string // 其他消息头信息
	Body() interface{}       // 解析后的消息体
	Data() *buffer.Buffer    // 原始数据
	Value(key string) string // 查询数据
	Codec() codec.Codec      // 消息体编解码
	// set操作
	SetReply(bool)
	SetStatus(uint)
	SetID(uint)
	SetName(string)
	SetSeqID(string)
	SetMethod(string)
	SetService(string)
	SetHead(map[string]string)
	SetBody(interface{})
	SetData(*buffer.Buffer)
	SetValue(key string, value string)
	SetCodec(codec.Codec)

	// 解析消息包体
	Parse(msg interface{}) error

	// 消息包头编解码
	Encode(b *buffer.Buffer) error
	Decode(b *buffer.Buffer) error
}

type NewPacket func() Packet
