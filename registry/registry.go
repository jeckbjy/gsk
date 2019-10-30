package registry

import "errors"

var (
	// Only can register one service
	ErrBadRegisterNum = errors.New("bad register num")
	// nodeId muse be unique and not empty
	ErrBadNodeId = errors.New("bad node id")
	// Not found error when GetService is called
	ErrNotFound = errors.New("not found")
	// Watcher stopped error when watcher is stopped
	ErrWatcherStopped = errors.New("watcher stopped")
	// Not ready, please wait
	ErrWait = errors.New("not ready, wait")
)

const (
	ActionCreate = "create"
	ActionUpdate = "update" // 更新状态
	ActionDelete = "delete"
)

// Wildcard 表示查询所有版本
const Wildcard = "*"

var registryMap = make(map[string]Registry)

func Add(r Registry) {
	registryMap[r.String()] = r
}

func Get(name string) Registry {
	return registryMap[name]
}

// Registry 服务注册与发现
type Registry interface {
	Init(...Option) error
	Options() *Options
	// 启动,某些库需要起一个定时器
	Start()
	// 退出,清理资源,同时需要释放注册的服务
	Stop() error
	// 注册服务,注册的是node的实例ID,一次只允许注册一个节点,同时实现库需要维护KeepAlive
	// 底层应该保证断线重连?
	Register(*Service, ...RegisterOption) error
	// 注销服务,注意注销的是node的实例ID
	Deregister(nodeID string) error
	// 查询服务,可以额外指定一个版本号,版本号为星号(*),表示查询所有版本
	Query(service string, version string) ([]*Service, error)
	// 罗列当前所有服务
	List() ([]*Service, error)
	// 创建一个Watcher,用于服务发现
	Watch(...WatchOption) (Watcher, error)
	String() string
}

// Watcher is an interface that returns updates
// about services within the registry.
// IWatcher首次创建回自动请求一次已经注册过的服务,然后主动触发Next事件,不需要外部再单独调用一次Query
type Watcher interface {
	// Add new service need watch
	Observe(services ...string)
	// Next is a blocking call
	Next() (*Result, error)
	Stop()
}

// Result is returned by a call to Next on
// the watcher. Actions can be create, update, delete
type Result struct {
	Action  string
	NodeID  string     // 任何操作都会有效
	Status  NodeStatus // 只有Update才有效
	Service *Service   // 只有Create才有效
}

func NewCreateResult(srv *Service) *Result {
	node := srv.First()
	node.Status = StatusAlive
	return &Result{Action: ActionCreate, NodeID: node.Id, Status: StatusAlive, Service: srv}
}

func NewUpdateResult(nodeId string, status NodeStatus) *Result {
	return &Result{Action: ActionUpdate, NodeID: nodeId, Status: status}
}

func NewDeleteResult(nodeId string, srv *Service) *Result {
	return &Result{Action: ActionDelete, NodeID: nodeId, Status: StatusDead, Service: srv}
}
