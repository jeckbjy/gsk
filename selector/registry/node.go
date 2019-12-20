package registry

import (
	"log"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/selector"
)

type _Node struct {
	srv  *registry.Service
	conn anet.Conn
	mux  sync.Mutex
}

func (n *_Node) Id() string {
	return n.srv.Id
}

func (n *_Node) Addr() string {
	return n.srv.Addr
}

func (n *_Node) Conn(tran anet.Tran) (anet.Conn, error) {
	var conn anet.Conn
	var err error
	n.mux.Lock()
	if n.conn == nil {
		conn, err = tran.Dial(n.Addr())
		n.conn = conn
		if conn == nil && err == nil {
			log.Printf("bad conn")
		}
	}
	n.mux.Unlock()
	return conn, err
}

func (n *_Node) Close() error {
	var err error
	n.mux.Lock()
	if n.conn != nil {
		err = n.conn.Close()
		n.conn = nil
	}
	n.mux.Unlock()
	return err
}

type _Group struct {
	nodes  []*_Node
	shadow []selector.Node
}

func (g *_Group) Add(node *_Node) {
	g.nodes = append(g.nodes, node)
	g.shadow = nil
}

func (g *_Group) Remove(id string) {
	for i, n := range g.nodes {
		if n.Id() == id {
			g.nodes = append(g.nodes[:i], g.nodes[i+1:]...)
		}
	}

	g.shadow = nil
}

func (g *_Group) Shadow() []selector.Node {
	if g.shadow != nil {
		return g.shadow
	}
	if len(g.nodes) == 0 {
		return nil
	}

	for _, n := range g.nodes {
		g.shadow = append(g.shadow, n)
	}

	return g.shadow
}

func (g *_Group) Filter(filters map[string]string) []selector.Node {
	results := make([]selector.Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		if n.srv.Match(filters) {
			results = append(results, n)
		}
	}

	return results
}
