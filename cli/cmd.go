package cli

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Command支持自动绑定字段
// 指针和数组默认类型为flag可选填,其他默认为必填参数且顺序唯一,除非使用flag tag指定类型
type _Command struct {
	name   string
	full   string
	group  string
	subs   []Command
	action Action
	fields []*_Field
	meta   map[string]string
}

func (c *_Command) Name() string {
	return c.name
}

func (c *_Command) Full() string {
	return c.full
}

func (c *_Command) Group() string {
	return c.group
}

func (c *_Command) Meta() map[string]string {
	return c.meta
}

func (c *_Command) Subs() []Command {
	return c.subs
}

func (c *_Command) Add(sub Command) {
	c.subs = append(c.subs, sub)
}

func (c *_Command) CanRun() bool {
	return c.action != nil
}

func (c *_Command) Run(ctx Context) error {
	if len(c.fields) == 0 {
		// 没有额外参数,无需绑定数据
		return c.action.Run(ctx)
	} else {
		action := reflect.New(reflect.TypeOf(c.action).Elem()).Interface().(Action)
		if err := c.Bind(ctx, action); err != nil {
			return err
		}
		return action.Run(ctx)
	}
}

// 解析Action的field信息
// 支持的关键字有:required,flag(h|help),def,min,max,desc
func (c *_Command) Parse(action Action) error {
	t := reflect.TypeOf(action).Elem()
	idx := 0
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("cli")
		if tag == "-" {
			// 不需要解析
			continue
		}

		if err := isSupportType(f.Type); err != nil {
			return err
		}

		field := &_Field{Index: -1, Type: f.Type, Kind: kindArgv}
		kind := f.Type.Kind()
		if kind == reflect.Ptr || kind == reflect.Slice {
			field.Kind = kindFlag
		}

		if err := field.fillTag(&f, tag); err != nil {
			return err
		}

		if field.Index == -1 && field.Kind == kindArgv {
			field.Index = idx
			idx++
		}
		c.fields = append(c.fields, field)
	}
	return nil
}

// 校验并绑定参数
func (c *_Command) Bind(ctx Context, action Action) error {
	v := reflect.ValueOf(action).Elem()
	for idx, info := range c.fields {
		field := v.Field(idx)
		switch info.Kind {
		case kindArgv:
			if err := bindArgv(ctx, field, info); err != nil {
				return err
			}
		case kindFlag:
			if err := bindFlag(ctx, field, info); err != nil {
				return err
			}
		case kindMeta:
			if err := bindMeta(ctx, field, info); err != nil {
				return err
			}
		}
	}

	return nil
}

func bindArgv(ctx Context, field reflect.Value, info *_Field) error {
	if info.Index >= ctx.NArg() {
		return fmt.Errorf("param not enough,len %+v, index %+v, name %+v", ctx.NArg(), info.Index, info.Name)
	}

	val := ctx.Arg(info.Index)
	return bindField(field, info, val)
}

func bindMeta(ctx Context, field reflect.Value, info *_Field) error {
	val := ctx.Get(info.MetaKey)
	if val == "" {
		val = info.Default
	}
	if val == "" {
		return fmt.Errorf("not found meta data,meta key %+v, field %+v", info.MetaKey, info.Name)
	}
	return bindValue(field, info, val)
}

func bindFlag(ctx Context, field reflect.Value, info *_Field) error {
	var values []string
	for _, k := range info.Flags {
		vv := ctx.Flag(k)
		if vv != nil {
			values = append(values, vv...)
		}
	}

	switch len(values) {
	case 0:
		if info.Default == "" {
			return bindValue(field, info, info.Default)
		} else if info.Required {
			return fmt.Errorf("%+v is required,not found data", info.Name)
		}
	case 1:
		return bindValue(field, info, values[0])
	default:
		if info.Type.Kind() != reflect.Slice {
			return fmt.Errorf("param not match,need %+v, but slice", info.Type.Kind())
		}

		slicev := reflect.New(info.Type)
		elemt := info.Type.Elem()

		for _, text := range values {
			elemv := reflect.New(elemt)
			if err := bindValue(elemv, info, text); err != nil {
				return err
			}
			slicev = reflect.Append(slicev, elemv)
		}
		field.Set(slicev)
		return nil
	}

	return nil
}

func bindField(field reflect.Value, info *_Field, text string) error {
	switch info.Type.Kind() {
	case reflect.Ptr:
		return bindValue(field.Elem(), info, text)
	case reflect.Slice:
		return fmt.Errorf("type not match,need slice, %+v", info.Name)
	default:
		return bindValue(field, info, text)
	}
}

func bindValue(value reflect.Value, info *_Field, text string) error {
	switch info.Type.Kind() {
	case reflect.Bool:
		x, err := strconv.ParseBool(text)
		if err != nil {
			return err
		}
		value.SetBool(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return err
		}
		if err := info.validateRange(float64(x)); err != nil {
			return err
		}
		value.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return err
		}
		if err := info.validateRange(float64(x)); err != nil {
			return err
		}
		value.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		if err := info.validateRange(x); err != nil {
			return err
		}
		value.SetFloat(x)
	case reflect.String:
		if err := info.validateIn(text); err != nil {
			return err
		}
		value.SetString(text)
	default:
		return ErrNotSupport
	}
	return nil
}

// 去除所有空格
func trimSpaceAll(tokens []string) []string {
	for i, t := range tokens {
		tokens[i] = strings.TrimSpace(t)
	}
	return tokens
}

// 判断是否支持
func isSupportType(t reflect.Type) error {
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	kind := t.Kind()
	if kind >= reflect.Bool && kind <= reflect.Float64 && kind != reflect.Uintptr {
		return nil
	}
	if kind == reflect.String {
		return nil
	}
	return errors.New("not support type")
}
