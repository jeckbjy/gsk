package selector

import (
	"github.com/jeckbjy/gsk/anet"
)

// Selector 客户端LoadBalance
// 默认基于Registry实现
// 真实的场景可能会更加复杂,比如Gateway或Proxy,根据不同的服务器需要采用不同的策略
// 比如通过UID或者GuildID,再或者活动ID等hash到不同的服务器
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
	if len(nodes) == 1 {
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
