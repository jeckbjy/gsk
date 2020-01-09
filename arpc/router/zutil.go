package router

import (
	"reflect"

	"github.com/jeckbjy/gsk/arpc"
)

// 用于粗略检测函数原型中参数是否是消息类型
// 要求:类型是结构体指针
func isMessage(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func isContext(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*arpc.Context)(nil)).Elem())
}

func isError(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
