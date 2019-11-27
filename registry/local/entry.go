package local

import (
	"container/heap"
	"time"

	"github.com/jeckbjy/gsk/registry"
)

func newEntry(srv *registry.Service, ttl int) *entry {
	return &entry{srv: srv, ttl: int64(ttl)}
}

type entry struct {
	srv    *registry.Service //
	ttl    int64             //
	expire int64             // 到期时间戳
}

func (e *entry) Reset() {
	e.expire = time.Now().UnixNano()/int64(time.Millisecond) + e.ttl
}

type eheap []*entry

func (h eheap) Len() int {
	return len(h)
}

func (h eheap) Less(i, j int) bool {
	return h[i].expire < h[j].expire
}

func (h eheap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *eheap) Push(x interface{}) {
	*h = append(*h, x.(*entry))
}

func (h *eheap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *eheap) Add(srv *registry.Service, ttl int) {
	e := newEntry(srv, ttl)
	e.Reset()
	heap.Push(h, e)
}

// 查询索引
func (h eheap) Find(srvId string) int {
	for i, e := range h {
		if e.srv.Id == srvId {
			return i
		}
	}

	return -1
}

// Update 更新heap位置
func (h *eheap) Update(srvId string) {
	if index := h.Find(srvId); index != -1 {
		e := (*h)[index]
		e.Reset()
		heap.Fix(h, index)
	}
}

func (h *eheap) Remove(srvId string) {
	if index := h.Find(srvId); index != -1 {
		heap.Remove(h, index)
	}
}
