package local

import (
	"container/heap"
	"errors"
	"sync"
	"time"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/ssdp"
)

var errHasRegistered = errors.New("has been registered")

// ssdp并没有提供过期通知,需要自己检查过期
type localWatcher struct {
	mux      sync.Mutex
	observed map[string]bool              // 需要监听的服务,空表示全部
	monitor  *ssdp.Monitor                //
	result   chan *registry.Result        //
	ticker   *time.Ticker                 // 定时器,用于检测超时
	services map[string]*registry.Service // 所有接收到的服务,用于快速查询
	expires  eheap                        // 需要检查超时的服务,小顶堆保存
	quit     chan bool                    //
	init     bool                         // 是否初始化过
}

func (w *localWatcher) Init(opts *registry.WatchOptions) error {
	w.init = true
	w.quit = make(chan bool, 1)
	w.result = make(chan *registry.Result, 1)
	w.observed = make(map[string]bool)
	w.services = make(map[string]*registry.Service)

	w.Observe(opts.Services...)

	w.monitor = &ssdp.Monitor{Alive: w.onAlive, Bye: w.onBye}
	if err := w.monitor.Start(); err != nil {
		return err
	}

	return nil
}

func (w *localWatcher) Observe(services ...string) {
	needQuery := false
	w.mux.Lock()
	for _, s := range services {
		if w.observed[s] == false {
			w.observed[s] = true
			needQuery = true
		}
	}
	w.mux.Unlock()

	if needQuery {
		go w.search()
	}
}

func (w *localWatcher) Next() (*registry.Result, error) {
	r, ok := <-w.result
	if !ok {
		return nil, registry.ErrWatcherStopped
	}

	return r, nil
}

func (w *localWatcher) Stop() {
	close(w.result)
	w.quit <- true
}

func (w *localWatcher) tick() {
	// start ticker
	w.ticker = time.NewTicker(time.Millisecond * 30)
	expires := make([]*registry.Service, 0)

loop:
	for {
		select {
		case <-w.ticker.C:
			expires = expires[:0]
			now := time.Now().UnixNano() / int64(time.Millisecond)

			w.mux.Lock()
			// process node expire
			for {
				if len(w.expires) == 0 || now < w.expires[0].expire {
					break
				}

				// expired,so remove and notify
				srv := heap.Pop(&w.expires).(*entry).srv
				nodeId := srv.Nodes[0].Id
				delete(w.services, nodeId)
				expires = append(expires, srv)
			}
			w.mux.Unlock()
			// notify
			for _, srv := range expires {
				w.result <- registry.NewDeleteResult(srv.First().Id, srv)
			}

		case <-w.quit:
			break loop
		}
	}
}

func (w *localWatcher) search() {
	results, err := ssdp.Search(serviceTarget, 1, "")
	if err != nil {
		return
	}

	services := make([]*registry.Result, 0)

	w.mux.Lock()
	for _, s := range results {
		srv, err := w.addService(s.USN, s.Location, s.Server, s.MaxAge())
		if srv != nil && err == nil {
			services = append(services, registry.NewCreateResult(srv))
		}
	}
	w.mux.Unlock()
	// 通知
	for _, s := range services {
		w.result <- s
	}
}

func (w *localWatcher) addService(nodeId string, name string, data string, ttl int) (*registry.Service, error) {
	// 并未关注,无需创建
	if len(w.observed) == 0 || w.observed[name] == false {
		return nil, nil
	}

	// 已经注册过了,不需要重复注册
	if srv, ok := w.services[nodeId]; ok {
		return srv, errHasRegistered
	}

	srv, err := registry.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	w.services[nodeId] = srv
	w.expires.Add(srv, ttl)

	return srv, nil
}

func (w *localWatcher) onAlive(m *ssdp.AliveMessage) {
	if m.Type != serviceTarget {
		return
	}

	w.mux.Lock()
	srv, err := w.addService(m.USN, m.Location, m.Server, m.MaxAge())
	if err == errHasRegistered {
		w.expires.Update(m.USN)
	}
	w.mux.Unlock()

	if err == nil && srv != nil {
		w.result <- registry.NewCreateResult(srv)
	}
}

func (w *localWatcher) onBye(m *ssdp.ByeMessage) {
	if m.Type != serviceTarget {
		return
	}

	var result *registry.Result
	w.mux.Lock()
	nodeId := m.USN
	if srv, ok := w.services[nodeId]; ok {
		delete(w.services, nodeId)
		w.expires.Remove(nodeId)
		result = registry.NewDeleteResult(m.USN, srv)
	}
	w.mux.Unlock()

	if result != nil {
		w.result <- result
	}
}
