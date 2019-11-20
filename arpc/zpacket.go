package arpc

import (
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

type ContentType int

const (
	CTProtoBuf ContentType = iota
	CTJson
	CTXml
	CTText
)

type CommandType int

const (
	CmdMsg CommandType = 0 // 正常消息通信
)

type HeadFlag uint

const (
	HFAck         HeadFlag = 0
	HFStatus               = 1
	HFContentType          = 2
	HFCommand              = 3
	HFSeqID                = 4
	HFMsgID                = 5
	HFName                 = 6
	HFMethod               = 7
	HFService              = 8
	HFHeadMap              = 9
	HFExtra                = 10
	HFMax                  = 15
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
//	1:系统依赖必须的字段,类型固定:比如Ack,Status,ContentType,Command,SequenceID,ID,Name,Service,
//	2:系统非必须但很常用,类型string:比如TraceID,SpanID,RemoteIP,UserID,Project,Auth,Checksum
//  3:Key-Value类型Head:
// Body:
// 	需要根据ContentType进行编解码,需要根据MsgID等信息查询到具体类型,因此解码需要分成两个接口
//  body需要是个指针类型
//
// Ack:是否是应答消息
// Status类似http的错误码,0表示OK
// ContentType使用枚举形式,默认protobuf和json
// Command:系统内控制命令,通常为0,表示消息通信,其他可用于HealthCheck等系统内预定义的命令
// SeqID:唯一序列号,用于RPC调用,全局唯一
// MsgID:消息静态唯一ID,不超过65535
// Name :消息名
// Method:调用方法名
// Service:服务类型,用于消息路由,也可以不使用此字段,而是自行根据消息ID分段或者自行编码
// Extra: 扩展字段,使用者可自行定义含义, 使用int索引定位,不能超过7
// Head:附加参数,kv结构,更加灵活,但是消耗也会更多,key要求不能含有|
type Packet interface {
	IsAck() bool
	SetAck(ack bool)
	Status() uint
	SetStatus(status uint)
	ContentType() ContentType
	SetContentType(ct ContentType)
	Command() CommandType
	SetCommand(CommandType)
	SeqID() string
	SetSeqID(id string)
	MsgID() uint16
	SetMsgID(id uint16)
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
	// 编解码接口
	Encode() error
	Decode() error
	// 解析body
	DecodeBody(msg interface{}) error
}

type NewPacket func() Packet
