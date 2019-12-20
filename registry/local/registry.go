package local

import (
	"container/heap"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/errorx"
)

const (
	defaultRootDir = ".services" //
	defaultTTL     = 10 * 1000   // 单位毫秒
)

func New() registry.Registry {
	r := &localRegistry{
		root:     defaultRootDir,
		ttl:      defaultTTL,
		services: make(map[string]*localEntry),
		quit:     make(chan bool),
	}
	return r
}

// 使用本地文件夹作为注册发现
type localRegistry struct {
	mux      sync.Mutex
	services map[string]*localEntry // 所有本地注册的服务
	alive    localHeap              // 用于定时注册
	watcher  *localWatcher          // watch,Lazy start
	ticker   *time.Ticker           // 定时检测, Lazy start
	root     string                 // 服务根目录
	ttl      int                    // 过期时间
	quit     chan bool              // 退出信号
}

func (r *localRegistry) Name() string {
	return "local"
}

func (r *localRegistry) Flush(e *localEntry, remove bool) error {
	srv := e.Srv
	filename := filepath.Join(r.root, encodeID(srv.Id, srv.Name))
	if remove {
		return os.Remove(filename)
	} else {
		e.Expired = time.Now().UnixNano()/int64(time.Millisecond) + int64(e.Ttl)
		data := e.Marshal()
		return ioutil.WriteFile(filename, data, os.ModePerm)
	}
}

func (r *localRegistry) Register(service *registry.Service) error {
	r.mux.Lock()

	// lazy start
	needRun := false
	if r.ticker == nil {
		needRun = true
		r.ticker = time.NewTicker(time.Millisecond * 30)
		_ = os.MkdirAll(r.root, os.ModePerm)
	}

	entry, ok := r.services[service.Id]
	if !ok {
		entry = &localEntry{Srv: service, Ttl: r.ttl}
		entry.Reset()
		heap.Push(&r.alive, entry)
		r.services[service.Id] = entry
	}

	r.mux.Unlock()

	if needRun {
		go r.tick()
	}

	return r.Flush(entry, false)
}

func (r *localRegistry) Unregister(serviceID string) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if entry, ok := r.services[serviceID]; ok {
		delete(r.services, serviceID)
		r.alive.Remove(serviceID)
		return r.Flush(entry, true)
	}

	return errorx.ErrNotFound
}

func (r *localRegistry) Query(name string, filters map[string]string) ([]*registry.Service, error) {
	// 遍历目录
	now := time.Now().UnixNano() / int64(time.Millisecond)
	results := make([]*registry.Service, 0)
	err := filepath.Walk(r.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		entry := &localEntry{}
		if err := entry.Unmarshal(data); err != nil {
			return err
		}

		// 过期了,删除
		if now > entry.Expired {
			_ = os.Remove(path)
			return nil
		}

		if name != "" && entry.Srv.Name != name {
			return nil
		}

		if !entry.Srv.Match(filters) {
			return nil
		}

		results = append(results, entry.Srv)

		return nil
	})

	return results, err
}

func (r *localRegistry) Watch(names []string, cb registry.Callback) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.watcher == nil {
		if err := os.MkdirAll(r.root, os.ModePerm); err != nil {
			return err
		}

		watcher := &localWatcher{}
		if err := watcher.Start(r.root); err != nil {
			return err
		}
		r.watcher = watcher
	}

	r.watcher.Add(names, cb)
	return nil
}

func (r *localRegistry) Close() error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.watcher != nil {
		r.watcher.Stop()
		r.watcher = nil
	}

	if r.ticker != nil {
		r.ticker.Stop()
		r.ticker = nil
	}

	// clear
	for _, e := range r.services {
		_ = r.Flush(e, true)
	}
	r.services = nil
	r.alive = nil

	r.tryRemoveRootDir()

	return nil
}

func (r *localRegistry) tryRemoveRootDir() {
	// remove empty dir
	count := 0
	filepath.Walk(r.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			count++
			return filepath.SkipDir
		}

		return nil
	})

	if count > 0 {
		_ = os.Remove(r.root)
	}
}

// 定时注册
func (r *localRegistry) tick() {
LOOP:
	for {
		select {
		case <-r.ticker.C:
			// 定时自动注册
			now := time.Now().UnixNano() / int64(time.Millisecond)
			max := r.alive.Len()
			for i := 0; i < max; i++ {
				if now < r.alive[0].Expired {
					break
				}
				e := r.alive[0]
				_ = r.Flush(e, false)
				e.Reset()
				heap.Fix(&r.alive, 0)
			}
		case <-r.quit:
			break LOOP
		}
	}
}
