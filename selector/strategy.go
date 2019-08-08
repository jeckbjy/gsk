package selector

import (
	"math/rand"
	"sync"
	"time"

	"github.com/jeckbjy/micro/registry"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetNodes(services []*registry.Service) []*registry.Node {
	if len(services) == 1 {
		return services[0].Nodes
	} else {
		var nodes []*registry.Node
		for _, service := range services {
			nodes = append(nodes, service.Nodes...)
		}
		return nodes
	}
}

// Random is a random strategy algorithm for node selection
func Random(services []*registry.Service) Next {
	nodes := GetNodes(services)
	if len(nodes) == 0 {
		return nil
	}

	return func() (*registry.Node, error) {
		i := rand.Int() % len(nodes)
		return nodes[i], nil
	}
}

// RoundRobin is a roundrobin strategy algorithm for node selection
func RoundRobin(services []*registry.Service) Next {
	nodes := GetNodes(services)
	if len(nodes) == 0 {
		return nil
	}

	var i = rand.Int()
	var mtx sync.Mutex

	return func() (*registry.Node, error) {
		mtx.Lock()
		node := nodes[i%len(nodes)]
		i++
		mtx.Unlock()

		return node, nil
	}
}
