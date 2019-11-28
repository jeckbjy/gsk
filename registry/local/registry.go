package local

import (
	"container/heap"
	"sync"
	"time"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/errorx"
	"github.com/jeckbjy/gsk/util/ssdp"
)

const serviceTarget = "_services"
const waitSec = 10

func New(opts *registry.Options) registry.Registry {
	r := &_Registry{}
	r.Init(opts)
	return r
}

type _Registry struct {
	mux         sync.Mutex                  //
	quit        chan bool                   // 退出信号
	ticker      *time.Ticker                // 定时检测过期服务,定时自动注册
	alive       eheap                       // 需要注册的服务,需要keepalive
	advertisers map[string]*ssdp.Advertiser // 所有注册的Id
	watchers    map[string]*_Watcher        // 所有监听的服务
	maxAge      int                         // 过期时间
	interval    int                         // 刷新间隔
}

func (r *_Registry) Name() string {
	return "local"
}

func (r *_Registry) Init(opts *registry.Options) {
	if opts != nil {
		r.maxAge = int(opts.TTL / time.Millisecond)
		r.interval = int(opts.Interval / time.Millisecond)
	} else {
		r.maxAge = 30
		r.interval = 15
	}

	r.advertisers = make(map[string]*ssdp.Advertiser)
	r.quit = make(chan bool, 1)

	// TODO: ssdp setup only use one?
	//en0, err := net.InterfaceByName("en0")
	//if err != nil {
	//	panic(err)
	//}
	//
	//ssdp.Interfaces = []net.Interface{*en0}
	go r.tick()
}

func (r *_Registry) Register(srv *registry.Service) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if _, ok := r.advertisers[srv.Id]; !ok {
		data, err := srv.Marshal()
		if err != nil {
			return err
		}

		ad, err := ssdp.Advertise(serviceTarget, srv.Id, srv.Name, data, r.maxAge)
		if err != nil {
			return err
		}

		if err := ad.Alive(); err != nil {
			return err
		}

		r.alive.Add(srv, r.interval)
		r.advertisers[srv.Id] = ad
	}

	return nil
}

// 根据服务ID注销
func (r *_Registry) Unregister(srvID string) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if ad, ok := r.advertisers[srvID]; ok {
		delete(r.advertisers, srvID)
		r.alive.Remove(srvID)
		return ad.Bye()
	}

	return errorx.ErrNotFound
}

func (r *_Registry) Query(name string, filters map[string]string) ([]*registry.Service, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	services, err := ssdp.Search(serviceTarget, waitSec, "")
	if err != nil {
		return nil, err
	}

	results := make([]*registry.Service, 0)
	for _, s := range services {
		srvName := s.Location
		if len(name) != 0 && name != srvName {
			continue
		}

		srv, err := registry.Unmarshal(s.Server)
		if err != nil {
			continue
		}

		// 筛选
		if filters != nil && len(filters) > 0 {
			if !srv.Match(filters) {
				continue
			}
		}

		results = append(results, srv)
	}

	return results, nil
}

func (r *_Registry) Watch(names []string, cb registry.Callback) error {
	r.mux.Lock()
	if r.watchers == nil {
		r.watchers = make(map[string]*_Watcher)
	}
	w, err := newWatcher(names, cb)
	if err == nil {
		r.watchers[w.Id] = w
	}
	r.mux.Unlock()
	return err
}

func (r *_Registry) Close() error {
	r.mux.Lock()
	// 自动注销所有服务
	for nodeID, ad := range r.advertisers {
		delete(r.advertisers, nodeID)
		r.alive.Remove(nodeID)
		_ = ad.Bye()
	}
	// 关闭所有监听
	for _, w := range r.watchers {
		_ = w.Close()
	}
	r.mux.Unlock()
	r.quit <- true

	return nil
}

func (r *_Registry) tick() {
	r.ticker = time.NewTicker(time.Millisecond * 30)
loop:
	for {
		select {
		case <-r.ticker.C:
			now := time.Now().UnixNano() / int64(time.Millisecond)
			r.mux.Lock()
			// process watcher
			for _, w := range r.watchers {
				w.tick(now)
			}
			// process alive
			r.keepAlive(now)

			r.mux.Unlock()
		case <-r.quit:
			break loop
		}
	}
}

func (r *_Registry) keepAlive(now int64) {
	count := 0
	for {
		if count >= len(r.alive) || now < r.alive[0].expire {
			break
		}
		count++

		e := r.alive[0]
		if ad, ok := r.advertisers[e.srv.Id]; ok {
			_ = ad.Alive()
		}

		r.alive[0].Reset()
		heap.Fix(&r.alive, 0)
	}
}
