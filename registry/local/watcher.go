package local

import (
	"container/heap"
	"sync"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/idgen/xid"
	"github.com/jeckbjy/gsk/util/ssdp"
)

func newWatcher(names []string, cb registry.Callback) (*_Watcher, error) {
	w := &_Watcher{}
	if err := w.Start(names, cb); err != nil {
		return nil, err
	}

	return w, nil
}

type _Watcher struct {
	monitor  *ssdp.Monitor
	mux      sync.Mutex
	Id       string
	observed map[string]bool   // 需要监听的服务,用于筛选服务
	nodes    map[string]bool   // 当前收到的服务,用于判断是否需要触发回调
	expired  eheap             // 用于检测过期
	cb       registry.Callback // 回调
}

func (w *_Watcher) Start(names []string, cb registry.Callback) error {
	w.nodes = make(map[string]bool)
	if names != nil {
		w.observed = make(map[string]bool)
		for _, n := range names {
			w.observed[n] = true
		}
	}
	w.cb = cb
	m := &ssdp.Monitor{Alive: w.onAlive, Bye: w.onBye}
	if err := m.Start(); err != nil {
		return err
	}
	w.monitor = m
	w.Id = xid.New().String()
	return w.query()
}

func (w *_Watcher) Close() error {
	if w.monitor != nil {
		return w.monitor.Close()
	}

	return nil
}

// 主动查询一次
func (w *_Watcher) query() error {
	services, err := ssdp.Search(serviceTarget, waitSec, "")
	if err != nil {
		return err
	}
	for _, s := range services {
		if w.observed != nil && w.observed[s.Location] == false {
			continue
		}

		w.add(s.USN, s.Server, s.MaxAge())
	}

	return nil
}

func (w *_Watcher) onAlive(m *ssdp.AliveMessage) {
	if m.Type != serviceTarget {
		return
	}
	serviceId := m.USN
	w.mux.Lock()
	if _, ok := w.nodes[serviceId]; !ok {
		w.nodes[serviceId] = true
		w.add(serviceId, m.Server, m.MaxAge())
	}
	w.mux.Unlock()
}

func (w *_Watcher) onBye(m *ssdp.ByeMessage) {
	if m.Type != serviceTarget {
		return
	}
	srvID := m.USN
	w.mux.Lock()
	w.expired.Remove(srvID)
	delete(w.nodes, srvID)
	w.mux.Unlock()
	ev := &registry.Event{Type: registry.EventDelete, Id: srvID}
	w.cb(ev)
}

// ssdp并没有提供过期通知,需要自己检查过期
func (w *_Watcher) tick(now int64) {
	w.mux.Lock()
	defer w.mux.Unlock()

	for {
		if len(w.expired) == 0 || now < w.expired[0].expire {
			break
		}
		srv := heap.Pop(&w.expired).(*entry).srv
		delete(w.nodes, srv.ID())
		w.cb(&registry.Event{Type: registry.EventDelete, Id: srv.ID()})
	}
}

func (w *_Watcher) add(srvId string, data string, ttl int) {
	srv, err := registry.Unmarshal(data)
	if err == nil {
		ev := &registry.Event{Type: registry.EventCreate, Id: srvId, Service: srv}
		w.cb(ev)
		w.expired.Add(srv, ttl)
		w.nodes[srvId] = true
	}
}
