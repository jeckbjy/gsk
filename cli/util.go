package cli

import (
	"encoding/csv"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func parseCommandLine(param string, comma rune) ([]string, error) {
	r := csv.NewReader(strings.NewReader(param))
	r.Comma = comma
	return r.Read()
}

type _Args struct {
	Params  []string
	Options map[string][]string
}

func (a *_Args) AddParam(value string) {
	a.Params = append(a.Params, value)
}

func (a *_Args) AddOption(key string, val string) error {
	if a.Options == nil {
		a.Options = make(map[string][]string)
	}

	if d, ok := a.Options[key]; ok {
		if len(val) == 0 || len(d) == 0 {
			// multiple 需要有参数
			return errors.New("multiple options need param")
		}
		a.Options[key] = append(d, val)
	} else {
		if len(val) == 0 {
			a.Options[key] = nil
		} else {
			a.Options[key] = []string{val}
		}
	}

	return nil
}

//
func parseArgs(args []string) (*_Args, error) {
	result := &_Args{}
	for idx := 0; idx < len(args); idx++ {
		token := args[idx]
		if token[0] != '-' {
			result.AddParam(token)
			continue
		}

		// parse flag
		if len(token) == 1 {
			// - 只能是最后一个
			if idx < len(args)-1 {
				return nil, errors.New("bad -,not the last")
			} else {
				_ = result.AddOption("-", "")
				return result, nil
			}
		}

		var short bool
		var key string
		var val string
		if token[1] == '-' {
			short = false
			key = token[2:]
		} else {
			short = true
			key = token[1:]
		}

		// parse flag value
		if strings.ContainsRune(key, '=') {
			// check has =
			values := strings.SplitAfterN(key, "=", 2)
			key = values[0]
			val = values[1]
		} else if idx+1 < len(args) && args[idx+1][0] != '-' {
			// check next
			idx++
			val = args[idx]
		}

		if short && len(key) > 1 {
			for i := 0; i < len(key)-1; i++ {
				if err := result.AddOption(string(key[i]), ""); err != nil {
					return nil, err
				}
			}
		} else {
			if err := result.AddOption(key, val); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// 将string转化成普通类型
func bindValue(str string, value interface{}) error {
	v := reflect.ValueOf(value)
	t := v.Type()
	if t.Kind() != reflect.Ptr {
		return errors.New("bind arg must be ptr")
	}

	arg := str
	switch t.Elem().Kind() {
	case reflect.String:
		v.SetString(arg)
	case reflect.Bool:
		x, err := strconv.ParseBool(arg)
		if err != nil {
			return err
		}
		v.SetBool(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(arg, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return err
		}
		v.SetFloat(x)
	default:
		return errors.New("bind arg not support")
	}

	return nil
}
