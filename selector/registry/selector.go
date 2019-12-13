package registry

import (
	"sync"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/selector"
	"github.com/jeckbjy/gsk/util/errorx"
)

func New(reg registry.Registry) selector.Selector {
	s := &_Selector{reg: reg, groups: make(map[string]*_Group), nodes: make(map[string]*_Node)}
	return s
}

type _Selector struct {
	mux    sync.Mutex
	reg    registry.Registry
	groups map[string]*_Group
	nodes  map[string]*_Node
}

func (s *_Selector) Name() string {
	return "registry"
}

func (s *_Selector) Select(service string, opts *selector.Options) (selector.Next, error) {
	g, err := s.getGroup(service)
	if err != nil {
		return nil, err
	}

	s.mux.Lock()
	var nodes []selector.Node
	if len(opts.Filters) > 0 {
		// filter每次都会拷贝,比较低效,不如定制selector,使用cache,慎重使用
		nodes = g.Filter(opts.Filters)
	} else {
		nodes = g.Shadow()
	}
	s.mux.Unlock()

	if len(nodes) == 0 {
		return nil, errorx.ErrNotAvailable
	}

	return opts.GetNext(nodes), nil
}

func (s *_Selector) getGroup(service string) (*_Group, error) {
	s.mux.Lock()
	if g, ok := s.groups[service]; ok {
		s.mux.Unlock()
		return g, nil
	}

	services, err := s.reg.Query(service, nil)
	if err != nil {
		s.mux.Unlock()
		return nil, err
	}

	g := &_Group{}
	s.groups[service] = g
	for _, srv := range services {
		s.upsert(srv)
	}
	s.mux.Unlock()

	if err := s.reg.Watch([]string{service}, s.onEvent); err != nil {
		return nil, err
	}

	return g, nil
}

func (s *_Selector) Close() error {
	s.mux.Lock()
	for _, n := range s.nodes {
		_ = n.Close()
	}
	s.nodes = make(map[string]*_Node)
	s.groups = make(map[string]*_Group)
	s.mux.Unlock()
	return nil
}

func (s *_Selector) onEvent(ev *registry.Event) {
	s.mux.Lock()
	switch ev.Type {
	case registry.EventUpsert:
		s.upsert(ev.Service)
	case registry.EventDelete:
		s.remove(ev.Id)
	}
	s.mux.Unlock()
}

func (s *_Selector) upsert(srv *registry.Service) {
	node, ok := s.nodes[srv.Id]
	if !ok {
		s.groups[srv.Name].Add(&_Node{srv: srv})
	} else if node.srv.Addr != srv.Addr {
		// ip change,maybe have some error?
		_ = node.Close()
	}
}

func (s *_Selector) remove(id string) {
	node, ok := s.nodes[id]
	if ok {
		_ = node.Close()
		name := node.srv.Name
		s.groups[name].Remove(id)
		delete(s.nodes, id)
	}
}
