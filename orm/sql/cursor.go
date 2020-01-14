package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

// https://medium.com/@ferencfbin/golang-own-structscan-method-for-sql-rows-978c5c80f9b5
// https://kylewbanks.com/blog/query-result-to-map-in-golang
type sqlCursor struct {
	rows     *sql.Rows
	columns  []interface{}
	pointers []interface{}
	dict     map[string]interface{}
}

func (c *sqlCursor) init() error {
	if c.dict != nil {
		return nil
	}

	cols, err := c.rows.Columns()
	if err != nil {
		return err
	}

	if len(cols) == 0 {
		return errors.New("no data")
	}

	c.columns = make([]interface{}, len(cols))
	c.pointers = make([]interface{}, len(cols))
	for i, _ := range c.columns {
		c.pointers[i] = &c.columns[i]
	}

	return nil
}

func (c *sqlCursor) Close() error {
	return c.rows.Close()
}

func (c *sqlCursor) Next() bool {
	return c.rows.Next()
}

func (c *sqlCursor) Scan(dest ...interface{}) error {
	return c.rows.Scan(dest...)
}

// 通过反射解析到struct中
func (c *sqlCursor) Decode(model interface{}) error {
	var v reflect.Value
	if value, ok := model.(reflect.Value); ok {
		v = value
	} else {
		v = reflect.ValueOf(model)
	}

	if v.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value")
	}

	if err := c.init(); err != nil {
		return err
	}

	// need clear columns?

	if err := c.rows.Scan(c.pointers...); err != nil {
		return err
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		tag := t.Field(i).Tag.Get("db")
		if len(tag) == 0 {
			// lower case?
			tag = t.Field(i).Name
		}
		item, ok := c.dict[tag]
		if !ok || item == nil {
			continue
		}
		// struct??
		// bind value
		// https://stackoverflow.com/questions/20767724/converting-unknown-interface-to-float64-in-golang
		v := reflect.ValueOf(item)
		v = reflect.Indirect(v)
		if !v.Type().ConvertibleTo(field.Type()) {
			return fmt.Errorf("cannot convert data")
		}
		fv := v.Convert(field.Type())
		field.Set(fv)
	}

	return nil
}
