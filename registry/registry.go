package registry

import "sync/atomic"

var defaultRegistry atomic.Value

func Default() Registry {
	return defaultRegistry.Load().(Registry)
}

func SetDefault(d Registry) {
	defaultRegistry.Store(d)
}

// 服务注册与发现
// Register:注册服务,底层会保持KeepAlive
// Unregister:根据服务ID注销服务
// Query:根据服务名查询服务,name空则表示全部,filter多个key表示且的关系
// Watch:监听服务,nil则监听全部,可同时监听多个服务,watch底层不会获取历史信息,需要主动Query一次
// Close:会自动Unregister所有服务,并关闭所有Watcher
type Registry interface {
	Name() string
	Register(service *Service) error
	Unregister(serviceID string) error
	Query(name string, filters map[string]string) ([]*Service, error)
	Watch(names []string, cb Callback) error
	Close() error
}

const (
	EventUpsert EventType = iota
	EventDelete
)

type EventType int

func (t EventType) String() string {
	switch t {
	case EventUpsert:
		return "upsert"
	case EventDelete:
		return "delete"
	default:
		return "unknown"
	}
}

type Event struct {
	Id      string
	Type    EventType
	Service *Service
}

type Callback func(ev *Event)
