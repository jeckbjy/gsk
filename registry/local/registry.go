package local

import (
	"container/heap"
	"sync"
	"time"

	"github.com/jeckbjy/micro/registry"
	"github.com/jeckbjy/micro/util/ssdp"
)

func init() {

}

const serviceTarget = "_services"

func New() registry.IRegistry {
	return &localRegistry{opts: &registry.Options{}}
}

// localRegistry
type localRegistry struct {
	opts        *registry.Options
	mux         sync.Mutex
	quit        chan bool                   // 退出信号
	ticker      *time.Ticker                // 定时检测过期服务,定时自动注册
	alives      eheap                       // 需要注册的服务,需要keepalive
	advertisers map[string]*ssdp.Advertiser //
	waitSec     int                         // 查询阻塞时间
}

func (r *localRegistry) Init(opts ...registry.Option) error {
	r.advertisers = make(map[string]*ssdp.Advertiser)

	r.waitSec = 1
	r.opts.Init(opts...)
	r.quit = make(chan bool, 1)

	// TODO: ssdp setup only use one?
	//en0, err := net.InterfaceByName("en0")
	//if err != nil {
	//	panic(err)
	//}
	//
	//ssdp.Interfaces = []net.Interface{*en0}
	return nil
}

func (r *localRegistry) Options() *registry.Options {
	return r.opts
}

func (r *localRegistry) Register(srv *registry.Service, opts ...registry.RegisterOption) error {
	if len(srv.Nodes) != 1 {
		return registry.ErrNotFound
	}

	conf := registry.RegisterOptions{}
	conf.Init(opts...)

	r.mux.Lock()
	defer r.mux.Unlock()

	node := srv.Nodes[0]
	if node.Id == "" {
		return registry.ErrBadNodeId
	}

	if _, ok := r.advertisers[node.Id]; !ok {
		data, err := registry.Marshal(srv)
		if err != nil {
			return err
		}

		ad, err := ssdp.Advertise(serviceTarget, node.Id, srv.Name, data, int(conf.TTL/time.Millisecond))
		if err != nil {
			return err
		}

		if err := ad.Alive(); err != nil {
			return err
		}

		r.alives.Add(srv, int(conf.Interval/time.Millisecond))
		r.advertisers[node.Id] = ad
	}

	return nil
}

func (r *localRegistry) Deregister(nodeID string) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if ad, ok := r.advertisers[nodeID]; ok {
		delete(r.advertisers, nodeID)
		r.alives.Remove(nodeID)
		return ad.Bye()
	}

	return registry.ErrNotFound
}

func (r *localRegistry) Query(service string, version string) ([]*registry.Service, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	query, err := r.search()
	if err != nil {
		return nil, err
	}

	// 根据版本聚合
	versions := make(map[string]*registry.Service)

	for _, srv := range query {
		if srv.Name != service {
			continue
		}

		if version != registry.Wildcard && version != srv.Version {
			continue
		}

		s, ok := versions[srv.Version]
		if !ok {
			s = &registry.Service{
				Name:      srv.Name,
				Version:   s.Version,
				Meta:      srv.Meta,
				Endpoints: srv.Endpoints,
			}
			versions[srv.Version] = s
		}
		s.Nodes = append(s.Nodes, srv.Nodes...)
	}

	results := make([]*registry.Service, 0)
	for _, v := range versions {
		results = append(results, v)
	}

	return results, nil
}

func (r *localRegistry) List() ([]*registry.Service, error) {
	query, err := r.search()
	if err != nil {
		return nil, err
	}

	// 根据服务名聚合
	services := make(map[string]*registry.Service)
	for _, srv := range query {
		fullID := srv.FullID()
		s, ok := services[fullID]
		if !ok {
			s = &registry.Service{
				Name:      srv.Name,
				Version:   srv.Version,
				Meta:      srv.Meta,
				Endpoints: srv.Endpoints,
			}
			services[fullID] = s
		}

		s.Nodes = append(s.Nodes, srv.Nodes...)
	}

	results := make([]*registry.Service, 0, 0)
	for _, v := range services {
		results = append(results, v)
	}

	return results, nil
}

func (r *localRegistry) Watch(opts ...registry.WatchOption) (registry.IWatcher, error) {
	o := registry.WatchOptions{}
	o.Init(opts...)
	w := &localWatcher{}
	if err := w.Init(&o); err != nil {
		return nil, err
	}

	return w, nil
}

func (r *localRegistry) String() string {
	return "local"
}

func (r *localRegistry) Start() {
	go r.tick()
}

func (r *localRegistry) Stop() error {
	r.mux.Lock()

	for nodeID, ad := range r.advertisers {
		delete(r.advertisers, nodeID)
		r.alives.Remove(nodeID)
		_ = ad.Bye()
	}

	r.mux.Unlock()

	r.quit <- true
	return nil
}

func (r *localRegistry) tick() {
	r.ticker = time.NewTicker(time.Millisecond * 30)
	expired := make([]*ssdp.Advertiser, 0)

loop:
	for {
		select {
		case <-r.ticker.C:
			expired = expired[:0]
			r.mux.Lock()
			now := time.Now().UnixNano() / int64(time.Millisecond)
			// process alive
			count := 0
			for {
				if count >= len(r.alives) || now < r.alives[0].expire {
					break
				}
				count++

				e := r.alives[0]
				nodeID := e.srv.Nodes[0].Id
				if ad, ok := r.advertisers[nodeID]; ok {
					expired = append(expired, ad)
				}

				r.alives[0].Reset()
				heap.Fix(&r.alives, 0)
			}
			r.mux.Unlock()
			// process
			for _, ad := range expired {
				_ = ad.Alive()
			}
		case <-r.quit:
			break loop
		}
	}
}

func (r *localRegistry) search() ([]*registry.Service, error) {
	services, err := ssdp.Search(serviceTarget, r.waitSec, "")
	if err != nil {
		return nil, err
	}

	results := make([]*registry.Service, 0)
	for _, s := range services {
		srv, err := registry.Unmarshal(s.Server)
		if err != nil {
			continue
		}
		results = append(results, srv)
	}

	return results, nil
}
