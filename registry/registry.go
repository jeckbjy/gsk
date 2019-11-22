package registry

// 服务注册与发现
// Register
//	注册时每个服务只能有1个Node,相同服务只需注册1次,底层会KeepAlive
// Unregister
// 	根据服务ID(NodeId)注销服务
// Query
// 	根据服务名查询服务,name空则表示全部,filter多个key表示且的关系
// Watch
// 	监听服务,nil则监听全部,可同时监听多个服务
// Close
//	会自动Unregister所有服务,并关闭所有Watcher
type Registry interface {
	Name() string
	Register(srv *Service) error
	Unregister(srvID string) error
	Query(name string, filters map[string]string) ([]*Service, error)
	Watch(names []string, cb Callback) error
	Close() error
}

const (
	EventCreate EventType = iota
	EventUpdate
	EventDelete
)

type EventType int

func (t EventType) String() string {
	switch t {
	case EventCreate:
		return "create"
	case EventUpdate:
		return "update"
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
