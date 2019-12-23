package arpc

import (
	"reflect"

	"github.com/jeckbjy/gsk/anet"
)

// Handler 消息回调处理函数
type Handler func(ctx Context) error

// Middleware 中间件
type Middleware func(Handler) Handler

type Context interface {
	Init(conn anet.Conn, msg Packet)
	Free()                      // 释放,可用于pool回收
	Conn() anet.Conn            // 原始Socket
	Message() Packet            // 消息信息
	Send(msg interface{}) error // 发送消息,不关心返回结果
	NewPacket() Packet          // 用于创建Response
}

func IsContext(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*Context)(nil)).Elem())
}

// 用于粗略检测函数原型中参数是否是消息类型
// 要求是指针且是结构体
func IsMessage(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func IsError(t reflect.Type) bool {
	return t.Implements(errorType)
}

type ContextFactory func() Context

var gContextFactory ContextFactory

func SetDefaultContextFactory(fn ContextFactory) {
	gContextFactory = fn
}

func NewContext() Context {
	return gContextFactory()
}
