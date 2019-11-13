package store

import (
	"context"
	"errors"
)

var ErrNotSupport = errors.New("not support")

// Store kv storage
// 主要用途:配置文件管理
// 推荐使用etcd,拥有mvcc控制
// 可以是本地文件存储,也可以是分布式kv存储,如etcd,consul,zookeeper
// consul的value限制不超过512kb
//
// Key需要保持路径一样的形式,例如:path/to/key
// Get:通过Key查询是否有数据,没有数据返回nil
// List:通过前缀查询,空返回全部
// Put: 通过key写入数据
// Delete: 通过Key删除数据,如果有额外参数Prefix,则表示前缀匹配所有
// Exists: 通过key判断是否存在
// Watch:  监听key变化,如果有额外参数Prefix,则表示前缀匹配(监听目录)
//
// https://github.com/abronan/valkeyrie
// https://etcd.io/
// https://www.consul.io/docs/agent/kv.html
type Store interface {
	List(ctx context.Context, key string, opts ...Option) ([]*KV, error)
	Get(ctx context.Context, key string, opts ...Option) (*KV, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string, opts ...Option) error
	Exists(ctx context.Context, key string) (bool, error)
	Watch(ctx context.Context, key string, cb Callback, opts ...Option) error
}

type EventType int

const (
	PUT    EventType = 0
	DELETE EventType = 1
)

type KV struct {
	Key            string
	Value          []byte
	CreateRevision int64
	ModifyRevision int64
	Version        int64
}

type Event struct {
	Type EventType
	Data *KV
	Prev *KV
}

type Callback func(ev *Event)
