package local

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jeckbjy/gsk/registry"
)

func encodeID(id string, name string) string {
	return fmt.Sprintf("%s.%s", id, name)
}

func decodeID(key string) (string, string) {
	index := strings.LastIndexByte(key, '.')
	if index == -1 {
		return "", ""
	}

	return key[:index], key[index+1:]
}

type localEntry struct {
	Srv     *registry.Service `json:"srv"`
	Ttl     int               `json:"ttl"`
	Expired int64             `json:"expired"` // 过期时间戳
}

func (e *localEntry) Marshal() []byte {
	data, _ := json.Marshal(e)
	return data
}

func (e *localEntry) Unmarshal(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *localEntry) Reset() {
	e.Expired = time.Now().UnixNano()/int64(time.Millisecond) + int64(e.Ttl)
}

type localHeap []*localEntry

func (h localHeap) Len() int {
	return len(h)
}

func (h localHeap) Less(i, j int) bool {
	return h[i].Expired < h[j].Expired
}

func (h localHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *localHeap) Push(x interface{}) {
	*h = append(*h, x.(*localEntry))
}

func (h *localHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// 查询索引
func (h localHeap) Find(srvId string) int {
	for i, e := range h {
		if e.Srv.Id == srvId {
			return i
		}
	}

	return -1
}

func (h *localHeap) Remove(srvId string) {
	if index := h.Find(srvId); index != -1 {
		heap.Remove(h, index)
	}
}
