package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrInvalidType      = errors.New("invalid type")
	ErrNoData           = errors.New("no data")
	ErrNotFoundNameLine = errors.New("not found name line")
)

// Unmarshal 根据列头信息反射到数组中
// 结构体映射分为三种情况:
// 一:第一行代表字段名,通过字段名映射
// 二:外部指定每列字段名
// 三:均没有指定的情况下,按照字段顺序映射
// 例如：
// client_id,client_name,client_age
// int,string,int
// id,name,age
// 1,Jose,42
// 2,Daniel,26
// 3,Vincent,32
//
//type Client struct { // Our example struct, you can use "-" to ignore a field
//	Id      string `csv:"client_id"`
//	Name    string `csv:"client_name"`
//	Age     string `csv:"client_age"`
//	NotUsed string `csv:"-"`
//}
// clients := []*Client{}
// Unmarshal(data, &clients)
//
func Unmarshal(data []byte, v interface{}, opts ...Option) error {
	o := Options{}
	o.Init(opts...)

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		return ErrInvalidType
	}

	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return ErrNoData
	}

	if len(records) <= o.IgnoreLines {
		// 只有消息头,没有数据
		return nil
	}

	//
	var name []string
	if o.NameLine >= 0 {
		if o.NameLine >= len(records) {
			return ErrNotFoundNameLine
		}
		name = records[o.NameLine]
	} else if len(o.NameHead) > 0 {
		name = o.NameHead
	}

	elemType := t.Elem().Elem().Elem()
	fieldIndex := toIndexMap(name, elemType)
	if len(fieldIndex) == 0 {
		return ErrInvalidType
	}

	records = records[o.IgnoreLines:]
	sliceV := reflect.MakeSlice(t.Elem(), 0, len(records))
	for _, record := range records {
		value := reflect.New(elemType)
		for j := 0; j < len(fieldIndex); j++ {
			idx := fieldIndex[j]
			if idx == -1 || idx >= len(record) {
				continue
			}
			field := value.Elem().Field(j)
			text := record[idx]
			_ = setValue(field, text)
		}
		sliceV = reflect.Append(sliceV, value)
	}

	vv := reflect.ValueOf(v)
	vv.Elem().Set(sliceV)

	return nil
}

type Unmarshaler interface {
	UnmarshalCSV([]byte) error
}

func toIndexMap(fields []string, t reflect.Type) []int {
	fieldIndex := make([]int, t.NumField())

	if len(fields) > 0 {
		fieldMap := make(map[string]int, len(fields))
		for i, v := range fields {
			fieldMap[v] = i
		}
		// 使用名字映射
		for i := 0; i < t.NumField(); i++ {
			fieldIndex[i] = -1

			f := t.Field(i)
			tag := f.Tag.Get("csv")
			if tag == "-" {
				continue
			}

			if tag == "" {
				tag = f.Name
			}

			idx, ok := fieldMap[tag]
			if ok {
				fieldIndex[i] = idx
			}
		}
	} else {
		// 按照顺序
		for i := 0; i < t.NumField(); i++ {
			fieldIndex[i] = i
		}
	}

	return fieldIndex
}

func setValue(val reflect.Value, str string) error {
	switch val.Kind() {
	case reflect.String:
		val.SetString(str)
	case reflect.Bool:
		value, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		val.SetBool(value)
	case reflect.Int, reflect.Int32, reflect.Int64:
		// Parse the value to an int
		value, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		val.SetInt(value)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		val.SetUint(value)
	case reflect.Float32, reflect.Float64:
		// Parse the value to an float
		value, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		val.SetFloat(value)
	case reflect.Slice:
		if !isNumberType(val.Type().Elem().Kind()) {
			return fmt.Errorf("is not number type")
		}

		// TODO: remove empty??
		tkn := strings.FieldsFunc(str, func(r rune) bool {
			return r == '|' || r == ','
		})

		result := reflect.MakeSlice(val.Type(), len(tkn), len(tkn))
		for i, t := range tkn {
			t = strings.TrimSpace(t)
			e := result.Index(i)
			if err := setValue(e, t); err != nil {
				return err
			}
		}
		val.Set(result)
	case reflect.Map:
		if !isNumberType(val.Type().Elem().Kind()) {
			return fmt.Errorf("is not number type")
		}

		tkn := strings.FieldsFunc(str, func(r rune) bool {
			// 范围有点广
			return unicode.IsPunct(r)
		})

		if len(tkn)%2 != 0 {
			return fmt.Errorf("bad map len")
		}

		typ := val.Type()
		// 每次都是new?
		keyv := reflect.New(typ.Key())
		valv := reflect.New(typ.Elem())
		result := reflect.MakeMap(typ)
		for i := 0; i < len(tkn); i += 2 {
			if err := setValue(keyv.Elem(), strings.TrimSpace(tkn[i])); err != nil {
				return err
			}
			if err := setValue(valv.Elem(), strings.TrimSpace(tkn[i+1])); err != nil {
				return err
			}
			result.SetMapIndex(keyv.Elem(), valv.Elem())
		}

		val.Set(result)
	case reflect.Struct:
		if u, ok := val.Interface().(Unmarshaler); ok {
			return u.UnmarshalCSV([]byte(str))
		}
		//switch val.Interface().(type) {
		//case time.Time:
		//	// 固定格式,Date
		//	tm, err := time.Parse("2006-01-02 15:04:05", str)
		//	if err != nil {
		//		return err
		//	}
		//	val.Set(reflect.ValueOf(tm))
		//case Week:
		//	return parseWeek(str, val)
		//default:
		//	return errors.New("not support struct")
		//}
	default:
		return fmt.Errorf("not support:%+v", val.Kind())
	}

	return nil
}

func isNumberType(kind reflect.Kind) bool {
	if kind >= reflect.Int && kind <= reflect.Float64 && kind != reflect.Uintptr {
		return true
	}

	return false
}
