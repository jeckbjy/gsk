package driver

import (
	"errors"
	"reflect"
)

var (
	ErrNoDocument = errors.New("no document found")
)

// 反射数据
func Decode(cursor Cursor, result interface{}) error {
	defer cursor.Close()

	v := reflect.ValueOf(result)
	switch v.Type().Kind() {
	case reflect.Ptr:
		if cursor.Next() {
			return cursor.Decode(result)
		}

		return ErrNoDocument
	case reflect.Slice:
		slicev := v.Elem()
		elemt := slicev.Type().Elem()
		for cursor.Next() {
			elemp := reflect.New(elemt)
			if err := cursor.Decode(elemp.Interface()); err != nil {
				return err
			}
			slicev = reflect.Append(slicev, elemp.Elem())
		}
		v.Elem().Set(slicev)
		return nil
	default:
		return errors.New("query result must be ptr or slice")
	}
}

// ParseModel 解析Model,获得Column信息
func ParseModel(model interface{}) ([]Column, error) {
	return nil, nil
}

//func ToFloat(data interface{}) (float64, error) {
//	destType := reflect.TypeOf(float64(0))
//	v := reflect.ValueOf(data)
//	v = reflect.Indirect(v)
//	if !v.Type().ConvertibleTo(destType) {
//		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
//	}
//
//	fv := v.Convert(destType)
//	return fv.Float(), nil
//}
//
//func ToInt(data interface{}) (int64, error) {
//	destType := reflect.TypeOf(int64(0))
//	v := reflect.ValueOf(data)
//	v = reflect.Indirect(v)
//	if !v.Type().ConvertibleTo(destType) {
//		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
//	}
//
//	fv := v.Convert(destType)
//	return fv.Int(), nil
//}
//
//func ToUint(data interface{}) (uint64, error) {
//	destType := reflect.TypeOf(int64(0))
//	v := reflect.ValueOf(data)
//	v = reflect.Indirect(v)
//	if !v.Type().ConvertibleTo(destType) {
//		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
//	}
//
//	fv := v.Convert(destType)
//	return fv.Uint(), nil
//}
//
//func ToBool(data interface{}) (bool, error) {
//	destType := reflect.TypeOf(false)
//	v := reflect.ValueOf(data)
//	v = reflect.Indirect(v)
//	if !v.Type().ConvertibleTo(destType) {
//		return false, fmt.Errorf("cannot convert %v to float64", v.Type())
//	}
//
//	fv := v.Convert(destType)
//	return fv.Bool(), nil
//}
//
//func ToString(data interface{}) (string, error) {
//	destType := reflect.TypeOf("")
//	v := reflect.ValueOf(data)
//	v = reflect.Indirect(v)
//	if !v.Type().ConvertibleTo(destType) {
//		return "", fmt.Errorf("cannot convert %v to float64", v.Type())
//	}
//
//	fv := v.Convert(destType)
//	return fv.String(), nil
//}
