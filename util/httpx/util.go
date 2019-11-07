package httpx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// 根据编码格式自动将数据编码
func Encode(contentType string, data interface{}) ([]byte, error) {
	if data == nil {
		return nil, nil
	}

	switch d := data.(type) {
	case string:
		return []byte(d), nil
	case []byte:
		return d, nil
	}

	switch contentType {
	case TypeJSON:
		return json.Marshal(data)
	case TypeXML:
		return xml.Marshal(data)
	case TypeForm:
		uv, err := EncodeUrlValues(data)
		if err != nil {
			return nil, err
		}
		r := uv.Encode()
		return []byte(r), nil
	case TypeHTML, TypeText:
		// must be string or []byte
		return nil, ErrInvalidType
	default:
		return nil, ErrNotSupport
	}
}

// 根据编码格式解码
func Decode(contentType string, data []byte, result interface{}) error {
	if result == nil {
		return nil
	}

	if len(data) == 0 {
		return ErrNoData
	}

	v := reflect.ValueOf(result)
	t := v.Type()
	// 只能是Ptr,虽然map不用Ptr某些情况也可以传出数据,但json.Unmarshal要求也是必须是Ptr
	if t.Kind() != reflect.Ptr {
		return ErrInvalidType
	}

	switch v := result.(type) {
	case *string:
		*v = string(data)
		return nil
	case *[]byte:
		*v = data
		return nil
	}

	switch contentType {
	case TypeJSON:
		return json.Unmarshal(data, result)
	case TypeXML:
		return xml.Unmarshal(data, result)
	case TypeForm:
		values, err := url.ParseQuery(string(data))
		if err != nil {
			return err
		}

		return DecodeUrlValues(values, result)
	default:
		return ErrNotSupport
	}
}

// convert map[string]string to url.Values
func EncodeUrlValues(data interface{}) (url.Values, error) {
	switch data.(type) {
	case url.Values:
		return data.(url.Values), nil
	case map[string]string:
		m := data.(map[string]string)
		r := url.Values{}
		for k, v := range m {
			r.Add(k, v)
		}
		return r, nil
	case map[string]interface{}:
		m := data.(map[string]interface{})
		r := url.Values{}
		for k, v := range m {
			kind := reflect.TypeOf(v).Kind()
			if kind <= reflect.Float64 {
				r.Add(k, fmt.Sprintf("%+v", v))
			} else if kind == reflect.Slice {
				vv := reflect.ValueOf(v)
				for i := 0; i < vv.Len(); i++ {
					f := vv.Field(i)
					r.Add(k, fmt.Sprintf("%+v", f.Interface()))
				}
			} else {
				return nil, ErrNotSupport
			}
		}
	default:
		return nil, ErrNotSupport
	}

	return nil, nil
}

func DecodeUrlValues(values url.Values, result interface{}) error {
	switch r := result.(type) {
	case *url.Values:
		*r = values
	case url.Values:
		for k, v := range values {
			r[k] = v
		}
	case *map[string]string:
		m := make(map[string]string)
		for k, v := range values {
			m[k] = v[0]
		}
		*r = m
	case *map[string]interface{}:
		m := make(map[string]interface{})
		for k, v := range values {
			m[k] = v[0]
		}
		*r = m
	case map[string]string:
		for k, v := range values {
			r[k] = v[0]
		}
	case map[string]interface{}:
		for k, v := range values {
			r[k] = v[0]
		}
	default:
		return ErrNotSupport
	}

	return nil
}

func ParseContentType(content string) (string, string) {
	idx := strings.LastIndexByte(content, ';')
	if idx == -1 {
		return content, ""
	}

	contentType := content[0:idx]
	charset := content[idx+1:]
	tokens := strings.Split(charset, "=")
	if len(tokens) > 1 {
		return contentType, tokens[1]
	} else {
		// invalid charset?
		return contentType, charset
	}
}
