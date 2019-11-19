package bi

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

// 默认使用json编码,且字段名固定
func DefaultEncode(events []*Message) []byte {
	x := make([]M, len(events))
	for i, v := range events {
		x[i] = toMap(v)
	}

	r, err := json.Marshal(x)
	if err != nil {
		return nil
	}

	return r
}

func toMap(m *Message) M {
	if m.Params == nil {
		m.Params = make(M)
	}
	m.Params["bi_id"] = m.ID
	m.Params["bi_event"] = m.Event
	m.Params["bi_timestamp"] = m.Time.UnixNano() / int64(time.Millisecond)
	return m.Params
}

// xx_yy_zz
func toSnake(s string) string {
	b := strings.Builder{}
	lastUpper := -1
	lastByte := byte(0)
	for i, v := range []byte(s) {
		if v >= 'A' && v <= 'Z' {
			// check add _
			if b.Len() > 0 && lastByte != '_' && i-lastUpper > 1 {
				b.WriteByte('_')
			}
			b.WriteByte(v - 'A' + 'a')
			lastUpper = i
		} else {
			b.WriteByte(v)
		}
		lastByte = v
	}

	return b.String()
}

// 通过反射获取要发送的内容
func Reflect(s interface{}) (string, M, error) {
	v := reflect.ValueOf(s)
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return "", nil, ErrBadEvent
	}

	event := toSnake(t.Name())
	params := make(M)
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		vf := v.Field(i)
		tag := tf.Tag.Get("bi")
		if tag == "" {
			tag = toSnake(tf.Name)
		}
		params[tag] = vf.Interface()
	}

	return event, params, nil
}
