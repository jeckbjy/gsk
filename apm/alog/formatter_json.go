package alog

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

func NewJsonFormatter(layout string, indent string) (Formatter, error) {
	f := &JsonFormatter{id: NewFormatterID()}

	if err := f.Parse(layout); err != nil {
		return nil, err
	}

	return f, nil
}

// 格式配置,例如:time=%D message=%m level=%l
type JsonFormatter struct {
	Indent string
	Layout string
	id     int
	fields map[string]*Layout
}

func (f *JsonFormatter) ID() int {
	return f.id
}

func (f *JsonFormatter) Name() string {
	return "json"
}

func (f *JsonFormatter) Format(entry *Entry) ([]byte, error) {
	if f.fields == nil && f.Layout != "" {
		if err := f.Parse(f.Layout); err != nil {
			return nil, err
		}
	}

	data := make(map[string]interface{}, len(entry.Fields)+4)
	for k, v := range f.fields {
		data[k] = v.Format(entry)
	}

	for k, v := range entry.Fields {
		data[k] = v
	}

	b := bytes.Buffer{}
	encoder := json.NewEncoder(&b)
	encoder.SetIndent("", f.Indent)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (f *JsonFormatter) Parse(layout string) error {
	// time="%d" text=%t
	r := csv.NewReader(strings.NewReader(layout))
	r.Comma = ' '
	fields, err := r.Read()
	if err != nil {
		return err
	}
	for _, field := range fields {
		tokens := strings.SplitN(field, "=", 2)
		if len(tokens) != 2 {
			return fmt.Errorf("bad token")
		}
		if l, err := NewLayout(tokens[1]); err != nil {
			return err
		} else {
			f.fields[tokens[0]] = l
		}
	}

	return nil
}
