package ebus

import "reflect"

type Callback func(event IEvent)

type IEvent interface {
}

type IEventBus interface {
	Len(topic string) int
	Listen(topic string, once bool, priority int, cb Callback) uint
	Remove(topic string, id uint)
	Clear(topic string)
	Emit(topic string, event IEvent)
}

var gEventBus = New()

func Len(topic string) int {
	return gEventBus.Len(topic)
}

// Emit 调用函数
func Emit(ev IEvent, opts ...Option) {
	o := Options{}
	for _, fn := range opts {
		fn(&o)
	}

	if o.topic == "" {
		o.topic = reflect.TypeOf(ev).Elem().Name()
	}

	gEventBus.Emit(o.topic, ev)
}

// Listen 注册监听函数,原型必须是func(ev XXEvent)
func Listen(callback interface{}, opts ...Option) uint {
	o := Options{}
	for _, fn := range opts {
		fn(&o)
	}

	cb, topic := toCallback(callback)
	if o.topic == "" {
		o.topic = topic
	}

	return gEventBus.Listen(o.topic, o.once, o.priority, cb)
}

// Remove 注销
func Remove(topic string, id uint) {
	gEventBus.Remove(topic, id)
}

// Clear 清空所有
func Clear(topic string) {
	gEventBus.Clear(topic)
}

func toCallback(callback interface{}) (Callback, string) {
	if cb, ok := callback.(Callback); ok {
		return cb, ""
	}

	v := reflect.ValueOf(callback)
	t := v.Type()
	// 原型必须是:func(ev XXEvent)
	if t.Kind() != reflect.Func || t.NumIn() != 1 || t.NumOut() != 0 {
		return nil, ""
	}

	arg0 := t.In(0)
	if !arg0.Implements(reflect.TypeOf((*IEvent)(nil)).Elem()) {
		return nil, ""
	}

	topic := t.In(0).Elem().Name()

	return func(ev IEvent) {
		in := []reflect.Value{reflect.ValueOf(ev)}
		v.Call(in)
	}, topic
}
