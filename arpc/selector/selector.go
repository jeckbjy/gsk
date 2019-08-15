package selector

import (
	"errors"

	"github.com/jeckbjy/gsk/registry"
)

// ISelector 筛选服务节点,同时也需要维护Conn缓存,因为外部没有办法知道节点的消失,主动删除Conn
type ISelector interface {
	Options() Options
	Init(opts ...Option) error
	Select(service string, opts ...SelectOption) (Next, error)
	Close()
}

// Next is a function that returns the next node
// based on the selector's strategy
type Next func() (*registry.Node, error)

// Filter is used to filter a service during the selection process
type Filter func([]*registry.Service) []*registry.Service

// Strategy is a selection strategy e.g random, round robin
type Strategy func([]*registry.Service) Next

var (
	ErrNotFound      = errors.New("not found")
	ErrNoneAvailable = errors.New("none available")
)
