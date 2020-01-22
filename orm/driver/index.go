package driver

import (
	"fmt"
	"strings"
)

type Order int

const (
	Asc  Order = 0 // 默认升序
	Desc       = 1
)

type IndexKey struct {
	Name  string
	Order Order
}

// TODO:Partial Index https://blog.huoding.com/2016/04/28/510
type Index struct {
	Name       string
	Keys       []IndexKey
	Background bool
	Unique     bool
	Sparse     bool
}

func (i *Index) GenerateName() {
	if len(i.Name) != 0 {
		return
	}

	builder := strings.Builder{}
	for _, v := range i.Keys {
		if builder.Len() > 0 {
			builder.WriteString("_")
		}
		builder.WriteString(fmt.Sprintf("%s_%d", v.Name, v.Order))
	}

	i.Name = builder.String()
}
