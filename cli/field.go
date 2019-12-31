package cli

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jeckbjy/gsk/util/strx"
)

const (
	tagIndex    = "index"    // 索引顺序,默认按照变量顺序,需要保证唯一,从零开始,例如:index=0
	tagRequired = "required" // 用于标识flag是否必须有数据,意义在于参数可以乱序传入,同时又约束必须存在,通常并不需要这样
	tagFlag     = "flag"     // 用于标识可选,与meta互斥,格式flag=h|help,不写help则使用变量名
	tagMeta     = "meta"     // 用于标识从meta中获取数据,格式:meta=xxx或者meta,默认使用变量名
	tagDesc     = "desc"     // 描述信息
	tagDefault  = "default"  // 默认值,例如:default=10
	tagMin      = "min"      // 最小值,例如:min=1
	tagMax      = "max"      // 最大值,例如:max=100
	tagIn       = "in"       // 取值列表,例如:in=mode1|mode2
)

// field类型
const (
	kindArgv = 0 // 必须参数,默认
	kindFlag = 1 // 可选参数
	kindMeta = 2 // 元参数
)

// 所有非flag字段都是必须参数
type _Field struct {
	Name     string       // 字段名
	Type     reflect.Type // 反射信息
	Kind     int          // 类型标识
	Index    int          // 参数索引
	Required bool         // 是否必须,仅flag有效,默认false
	Flags    []string     // 可选项名称
	MetaKey  string       // 元数据索引名
	Desc     string       // 描述信息
	Default  string       // 默认参数,只有flag类型有效
	Min      *float64     // 限制最小值
	Max      *float64     // 限制最大值
	In       []string     // 可用范围
}

func (field *_Field) fillTag(f *reflect.StructField, tag string) error {
	snakeName := strx.ToSnake(f.Name)
	// 第一个参数代表名,默认使用snake_case
	tokens := trimSpaceAll(strings.Split(tag, ","))
	if len(tokens) == 0 {
		return nil
	}
	for _, t := range tokens {
		kv := trimSpaceAll(strings.SplitN(t, "=", 2))
		if len(kv) < 1 {
			return ErrInvalidTag
		}
		// flag允许不带参数,默认为字段名
		key := kv[0]
		if key != tagFlag && key != tagMeta && key != tagRequired && len(kv) != 2 {
			return ErrInvalidTag
		}

		switch key {
		case tagIndex:
			if v, err := strconv.Atoi(kv[1]); err == nil {
				field.Index = v
			} else {
				return err
			}
		case tagRequired:
			field.Required = true
		case tagFlag:
			field.Kind = kindFlag
			if len(kv) == 2 {
				field.Flags = trimSpaceAll(strings.Split(kv[1], "|"))
			} else {
				field.Flags = []string{snakeName}
			}
		case tagMeta:
			field.Kind = kindMeta
			if len(kv) == 2 {
				field.MetaKey = kv[1]
			} else {
				field.MetaKey = snakeName
			}
		case tagDesc:
			field.Desc = kv[1]
		case tagDefault:
			field.Default = kv[1]
		case tagMin:
			if v, err := strconv.ParseFloat(kv[1], 64); err == nil {
				field.Min = &v
			} else {
				return err
			}
		case tagMax:
			if v, err := strconv.ParseFloat(kv[1], 64); err == nil {
				field.Max = &v
			} else {
				return err
			}
		case tagIn:
			field.In = trimSpaceAll(strings.Split(kv[1], "|"))
		default:
			return fmt.Errorf("unknown tag,%+v", kv[0])
		}
	}

	return nil
}

func (field *_Field) validateRange(x float64) error {
	if field.Min != nil {
		if x < *field.Min {
			return fmt.Errorf("invalid data, %+v must be great then %+v", field.Name, *field.Min)
		}
	}

	if field.Max != nil {
		if x > *field.Max {
			return fmt.Errorf("invalid data, %+v must be less then %+v", field.Name, *field.Max)
		}
	}

	return nil
}

func (field *_Field) validateIn(x string) error {
	if len(field.In) == 0 {
		return nil
	}
	for _, s := range field.In {
		if s == x {
			return nil
		}
	}

	return fmt.Errorf("cannot find %+v in %+v ", x, field.Name)
}
