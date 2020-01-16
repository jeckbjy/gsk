package cache

import (
	"errors"
)

var ErrNotFound = errors.New("key not found")

// Cache 简单的缓存系统接口,没有过期等高级功能
// TODO:增加hook等功能,当缓存不存在时,去db查询
type Cache interface {
	Keys() []interface{}
	Values() []interface{}
	Has(key interface{}) bool
	Put(key, value interface{}) error
	Get(key interface{}) (interface{}, error)
	Remove(key interface{}) interface{}
	Clear()
	Len() int
}
