package local

import (
	"container/heap"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/fsnotify"
)

type localWatcher struct {
	mux      sync.Mutex                   //
	watcher  *fsnotify.Watcher            //
	global   registry.Callback            // 全局回调
	observed map[string]registry.Callback // 根据服务名回调,与全局回调互斥
	expired  localHeap                    // 用于检测是否过期,在正常的时间戳上再加上一个延迟
	ticker   *time.Ticker                 // 定时器
	quit     chan bool                    // 退出信号
}

func (w *localWatcher) Start(rootDir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(rootDir); err != nil {
		_ = watcher.Close()
		return err
	}

	w.observed = make(map[string]registry.Callback)
	w.watcher = watcher
	w.ticker = time.NewTicker(time.Millisecond * 30)
	w.quit = make(chan bool)
	go w.loop()
	return nil
}

func (w *localWatcher) Stop() {
	w.quit <- true
	w.mux.Lock()
	w.global = nil
	w.observed = nil
	w.expired = nil
	if w.ticker != nil {
		w.ticker.Stop()
		w.ticker = nil
	}
	if w.watcher != nil {
		_ = w.watcher.Close()
		w.watcher = nil
	}
	w.mux.Unlock()
}

func (w *localWatcher) Add(names []string, cb registry.Callback) {
	w.mux.Lock()
	if len(names) == 0 {
		w.global = cb
	} else {
		if w.observed == nil {
			w.observed = make(map[string]registry.Callback)
		}

		for _, name := range names {
			w.observed[name] = cb
		}
	}

	w.mux.Unlock()
}

func (w *localWatcher) Del(names []string) {
	w.mux.Lock()
	if len(names) == 0 {
		w.global = nil
	} else if w.observed != nil {
		for _, name := range names {
			delete(w.observed, name)
		}
	}
	w.mux.Unlock()
}

func (w *localWatcher) loop() {
LOOP:
	for {
		select {
		case ev := <-w.watcher.Events:
			switch {
			case ev.Op.IsWrite():
				w.onWrite(&ev)
			case ev.Op.IsRemove():
				id, name := decodeID(ev.Name)
				w.onRemove(id, name, true)
			case ev.Op.IsRename():
				id, name := decodeID(ev.Name)
				w.onRemove(id, name, true)
			}
		case <-w.ticker.C:
			// 检测是否有过期服务
			now := time.Now().UnixNano() / int64(time.Millisecond)
			max := w.expired.Len()
			for i := 0; i < max; i++ {
				if now < w.expired[0].Expired {
					break
				}
				srv := heap.Pop(&w.expired).(*localEntry).Srv
				w.onRemove(srv.Id, srv.Name, false)
			}
		case <-w.quit:
			break LOOP
		}
	}
}

func (w *localWatcher) onWrite(ev *fsnotify.Event) {
	data, err := ioutil.ReadFile(ev.Name)
	if err != nil {
		log.Printf("%s,%+v", ev.Name, err)
		return
	}

	entry := &localEntry{}
	if err := entry.Unmarshal(data); err != nil {
		log.Printf("unmarshal fail,%+v", err)
		return
	}
	srv := entry.Srv
	w.mux.Lock()
	cb := w.getCallback(srv.Name)
	if cb != nil {
		// 不存在则添加,延迟一些时间
		entry.Expired += int64(entry.Ttl / 3)
		// upsert:
		if index := w.expired.Find(srv.Id); index == -1 {
			// 新建
			heap.Push(&w.expired, entry)
			cb(&registry.Event{Id: ev.Name, Type: registry.EventUpsert, Service: entry.Srv})
		} else {
			// 更新时间戳
			w.expired[index] = entry
			heap.Fix(&w.expired, index)
		}
	}
	w.mux.Unlock()
}

func (w *localWatcher) onRemove(id, name string, needRemoveExpired bool) {
	w.mux.Lock()
	cb := w.getCallback(name)
	if cb != nil {
		cb(&registry.Event{Id: id, Type: registry.EventDelete})
	}
	if needRemoveExpired {
		w.expired.Remove(id)
	}
	w.mux.Unlock()
}

func (w *localWatcher) getCallback(name string) registry.Callback {
	if name == "" {
		return nil
	}

	if w.global != nil {
		return w.global
	}

	if cb, ok := w.observed[""]; ok {
		return cb
	}

	return nil
}
