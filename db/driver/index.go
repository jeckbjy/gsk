package driver

import (
	"fmt"
	"sort"
	"strings"
)

type OrderType int

const (
	Asc OrderType = 1
	Des           = 0
)

type IndexKey struct {
	Key   string
	Order OrderType
}

type Index struct {
	Name       string
	Indexes    []IndexKey
	Background bool
	Unique     bool
	Sparse     bool
}

// 规则:keys可以是string,[]string,map[string]int,[]IndexKey,
// 如果不知道Order,则默认按照升序(ascending)排列
// 如果是map,则按照key顺序排列
func ParseIndex(keys interface{}) ([]IndexKey, error) {
	switch keys.(type) {
	case []IndexKey:
		return keys.([]IndexKey), nil
	case string:
		return []IndexKey{{keys.(string), Asc}}, nil
	case []string:
		values := keys.([]string)
		result := make([]IndexKey, 0, len(values))
		for _, v := range values {
			result = append(result, IndexKey{v, Asc})
		}
		return result, nil
	case map[string]int:
		values := keys.(map[string]int)
		sortedKeys := make([]string, 0, len(values))
		for k, _ := range values {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		result := make([]IndexKey, 0, len(values))
		for _, k := range sortedKeys {
			result = append(result, IndexKey{k, OrderType(values[k])})
		}

		return result, nil
	default:
		return nil, fmt.Errorf("index:not support,%+v", keys)
	}
}

// 自动创建名字
func ParseIndexName(keys []IndexKey) string {
	builder := strings.Builder{}
	for _, v := range keys {
		if builder.Len() > 0 {
			builder.WriteString("_")
		}
		builder.WriteString(fmt.Sprintf("%s_%d", v.Key, v.Order))
	}

	return builder.String()
}
