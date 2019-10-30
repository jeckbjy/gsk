package sql

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/jeckbjy/gsk/db/driver"
)

func toWhere(cond driver.Cond) (string, error) {
	if cond == nil {
		// no condition, query all record
		return "", nil
	}

	builder := sqlBuilder{}
	builder.Write(spBlank, "WHERE")
	if err := buildWhere(&builder, cond, 0); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func buildWhere(b *sqlBuilder, cond driver.Cond, depth int) error {
	switch op := cond.Operator(); {
	case op > driver.TOK_EQ && op <= driver.TOK_NIN:
		//
		s := cond.(driver.ExprCond)
		b.Write(spBlank, "%s %s %s", s.Key(), sqlToken[op], toString(s.Value()))
	case op == driver.TOK_AND || op == driver.TOK_OR:
		s := cond.(driver.ListCond)
		if len(s.List()) == 0 {
			return nil
		}
		if depth > 0 {
			b.Write(spBlank, "(")
		}

		for i, l := range s.List() {
			if i > 0 {
				b.Write(spBlank, sqlToken[op])
			}

			if err := buildWhere(b, l, depth+1); err != nil {
				return err
			}
		}
		if depth > 0 {
			b.Write(spBlank, ")")
		}
	default:
		return errors.New("not support")
	}

	return nil
}

// 只能是普通类型?
func toString(data interface{}) string {
	var v reflect.Value
	if value, ok := data.(reflect.Value); ok {
		v = value
	} else {
		v = reflect.ValueOf(data)
	}

	switch kind := v.Type().Kind(); {
	case kind == reflect.String:
		return fmt.Sprintf("'%s'", v.String())
	case kind >= reflect.Bool && kind <= reflect.Float64 && kind != reflect.Uintptr:
		return fmt.Sprintf("%+v", v.Interface())
	case kind == reflect.Slice:
	// ??
	case kind == reflect.Ptr:
		return toString(v.Elem())
	default:
		return ""
	}
	return ""
}

var sqlToken = [...]string{
	driver.TOK_EQ:  "=",
	driver.TOK_NE:  "<>",
	driver.TOK_GT:  ">",
	driver.TOK_GTE: ">=",
	driver.TOK_LT:  "<",
	driver.TOK_LTE: "<=",
	driver.TOK_IN:  "IN",
	driver.TOK_NIN: "NOT IN",
	driver.TOK_AND: "AND",
	driver.TOK_OR:  "OR",
	driver.TOK_NOT: "NOT",
}
