package selector

import "github.com/jeckbjy/gsk/anet"

// 客户端Load Balancer
type Selector interface {
	Name() string
	Select(service string, opts *Options) (Next, error)
	Close() error
}

// Node节点
type Node interface {
	Id() string
	Addr() string
	Conn(tran anet.Tran) (anet.Conn, error)
}

type Next func() (Node, error)

type Strategy func([]Node) Next

type Options struct {
	Filters  map[string]string
	Strategy Strategy
	Hash     int64
}

func (o *Options) GetNext(nodes []Node) Next {
	if len(nodes) == 0 {
		return First(nodes)
	}

	if o.Strategy != nil {
		return o.Strategy(nodes)
	}

	if o.Hash > 0 {
		return Hash(nodes, uint64(o.Hash))
	}

	return Random(nodes)
}
