package arpc

import (
	"fmt"
	"strconv"

	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

var (
	// 可以外部启动前指定修改
	IDMin = -100  // 最小消息ID,用于指定系统ID范围
	IDMax = 65535 // 最大消息ID
)

type PacketFactory func() Packet

var gPacketFactory PacketFactory

func SetPacketFactory(fn PacketFactory) {
	gPacketFactory = fn
}

func NewPacket() Packet {
	return gPacketFactory()
}

type HeadFlag uint

const (
	HFAck         HeadFlag = 0  // 标识是否是消息应答,bool
	HFStatus               = 1  // 返回状态信息,空表示OK,string
	HFContentType          = 2  // 编码协议,0表示使用默认双方约定的协议,int
	HFSeqID                = 3  // RPC唯一ID,string,改用uint64?
	HFMsgID                = 4  // 静态唯一消息ID,非零值,负数表示系统消息,正数表示用户消息
	HFNameMethod           = 5  // 编码格式,以/开始表示Method,否则表示消息名
	HFService              = 6  // 调用服务名,string,和method合并成1个值?
	HFHeadMap              = 7  // 自定义消息头,map[string]string
	HFExtra                = 8  // 扩展字段,key<16
	HFMax                  = 15 // 最大可用位数
)

// 预定义extra枚举,外部可以自行定义
const (
	HFExtraTraceID   = 0
	HFExtraSpanID    = 1
	HFExtraRemoteIP  = 2
	HFExtraUserID    = 3
	HFExtraProjectID = 4
)

// 使用2个字节作为Flag标识,目前系统已经使用了9个,还剩7个可以自定义,取值范围[0-6]
// 服务器集群内经常使用的有TraceID，SpanID,RemoteIP,UserID,ProjectID等
// 客户端与服务器通信经常使用的有,Auth,Checksum
const HFExtraMax = 6
const HFExtraMask = ^uint16(1<<HFExtra - 1)

// 私有通信协议
// 编码格式:Flag[2byte]+Head+Body
// Flag: 固定两个字节,每位标识对应的head是否有数据
// Head:
//	1:系统依赖必须的字段,类型固定:比如Ack,Status,ContentType,SequenceID,ID,Name,Method,Service,
//	2:系统非必须但很常用,类型string:比如TraceID,SpanID,RemoteIP,UserID,Project,Auth,Checksum
//  3:Key-Value类型Head:
// Body:
// 	需要根据ContentType进行编解码,需要根据MsgID等信息查询到具体类型,因此解码需要分成两个接口
//  body需要是个指针类型
//
// Ack:是否是应答消息
// Status:错误信息,status line格式,例如 "200 OK"
// ContentType使用枚举形式,默认protobuf和json
// SeqID:唯一序列号,用于RPC调用,全局唯一
// MsgID:消息静态唯一ID,不超过65535
// Name :消息名
// Method:调用方法名
// Service:服务类型,用于消息路由,也可以不使用此字段,而是自行根据消息ID分段或者自行编码
// Extra: 扩展字段,使用者可自行定义含义, 使用int索引定位,不能超过7
// Head:附加参数,kv结构,更加灵活,但是消耗也会更多,key要求不能含有|
//
// 有几个特殊字段,不需要进行编码通信,仅仅用于系统内部调度
// Internal:用于系统扩展,可以透传任意数据
type Packet interface {
	Reset()
	IsAck() bool
	SetAck(ack bool)
	Code() int
	Status() string
	SetStatus(code int, info string)
	ContentType() int
	SetContentType(ct int)
	SeqID() string
	SetSeqID(id string)
	MsgID() int
	SetMsgID(id int)
	Name() string
	SetName(name string)
	Method() string
	SetMethod(string)
	Service() string
	SetService(service string)
	Extra(key uint) string
	SetExtra(key uint, value string) error
	Head(key string) string
	SetHead(key string, value string)
	Body() interface{}
	SetBody(interface{})
	Codec() codec.Codec
	SetCodec(codec.Codec)
	Buffer() *buffer.Buffer
	SetBuffer(b *buffer.Buffer)
	// 不需要序列化字段
	Internal() interface{}
	SetInternal(interface{})
	// 编解码接口
	Encode(data *buffer.Buffer) error
	Decode(data *buffer.Buffer) error
}

// util help
func DecodeBody(pkg Packet, msg interface{}) error {
	decoder := pkg.Codec()
	if decoder == nil {
		return ErrNoCodec
	}
	if err := decoder.Decode(pkg.Buffer(), msg); err != nil {
		return err
	}
	pkg.SetBody(msg)
	return nil
}

// 状态信息,使用空格分开,不单纯使用Code,因为字符串能返回一些调试信息
// 空表示OK
// Code 与http状态码保存一致
type Status struct {
	Code int
	Info string
}

func (s *Status) Encode() string {
	if s.Code == 0 {
		return ""
	}
	return fmt.Sprintf("%d %s", s.Code, s.Info)
}

func (s *Status) Decode(text string) {
	if len(text) == 0 {
		s.Code = 0
		s.Info = ""
		return
	}

	for i, ch := range text {
		if ch < '0' || ch > '9' {
			if i == 0 {
				s.Code = 0
				s.Info = text
			} else {
				s.Code, _ = strconv.Atoi(text[:i-1])
				s.Info = text[i:]
			}
			break
		}
	}
}

func IsValidID(id int) bool {
	return id >= IDMin && id < IDMax
}
