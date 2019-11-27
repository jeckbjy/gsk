package selector

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func First(nodes []Node) Next {
	return func() (Node, error) {
		return nodes[0], nil
	}
}

func Random(nodes []Node) Next {
	return func() (Node, error) {
		i := rand.Intn(len(nodes))
		return nodes[i], nil
	}
}

func RoundRobin(nodes []Node) Next {
	var i = rand.Int()
	var mtx sync.Mutex

	return func() (Node, error) {
		mtx.Lock()
		node := nodes[i%len(nodes)]
		i++
		mtx.Unlock()

		return node, nil
	}
}

func Hash(nodes []Node, key uint64) Next {
	id := key % uint64(len(nodes))
	return func() (Node, error) {
		return nodes[id], nil
	}
}
