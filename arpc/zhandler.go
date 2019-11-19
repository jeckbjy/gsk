package arpc

import (
	"context"
	"reflect"

	"github.com/jeckbjy/gsk/anet"
)

// Handler 消息回调处理函数
type Handler func(ctx Context) error

// Middleware 中间件
type Middleware func(Handler) Handler

type Context interface {
	context.Context
	Reset()                     // 重置数据
	Conn() anet.Conn            // 原始Socket
	Request() Packet            // 消息请求
	Response() Packet           // 消息应答
	Send(msg interface{}) error // 发送消息,不关心返回结果
}

func IsContext(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*Context)(nil)).Elem())
}

// 用于粗略检测函数原型中参数是否是消息类型
// 要求是指针且是结构体
func IsMessage(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func IsError(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
