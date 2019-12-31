package cli

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

type _Context struct {
	params []string            // 必须参数,安下标索引,已经去除了command名字信息
	flags  map[string][]string // 可选数据,value可以是数组
	metas  map[string]string   // 元数据信息,通过外部传入
	result interface{}         // 返回结果
}

func (c *_Context) NArg() int {
	return len(c.params)
}

func (c *_Context) Arg(index int) string {
	return c.params[index]
}

func (c *_Context) NFlag() int {
	return len(c.flags)
}

func (c *_Context) Flag(key string) []string {
	if c.flags != nil {
		return c.flags[key]
	}

	return nil
}

func (c *_Context) Get(key string) string {
	if c.metas != nil {
		return c.metas[key]
	}

	return ""
}

func (c _Context) Result() interface{} {
	return c.result
}

func (c *_Context) JSON(data interface{}) error {
	result, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.result = string(result)
	return nil
}

func (c *_Context) XML(value interface{}) error {
	result, err := xml.Marshal(value)
	if err != nil {
		return err
	}
	c.result = string(result)
	return nil
}

func (c *_Context) Text(format string, values ...interface{}) error {
	if len(values) > 0 {
		c.result = fmt.Sprintf(format, values...)
	} else {
		c.result = format
	}

	return nil
}

func (c *_Context) Any(value interface{}) error {
	c.result = value
	return nil
}
