package local

import (
	"fmt"
	"github.com/jeckbjy/micro/registry"
	"github.com/jeckbjy/micro/util/ssdp"
	"log"
	"os"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	ssdp.Logger = log.New(os.Stderr, "[SSDP] ", log.LstdFlags)

	// start registry
	t.Log("start register")

	r := &localRegistry{}
	_ = r.Init()
	r.Start()
	node := &registry.Node{Id: "aaa", Address: "127.0.0.1", Port: 9999}
	s := &registry.Service{Name: "test", Nodes: []*registry.Node{node}}
	_ = r.Register(s)

	t.Log("start watch")
	w, err := r.Watch(registry.WithWatchServices("test"))
	if err != nil {
		t.Error(err)
		return
	}

	// start watch
	go func() {
		for {
			r, err := w.Next()
			if err != nil {
				break
			}
			switch r.Action {
			case registry.ActionCreate:
				t.Logf("new service:%+v", r.Service.Nodes[0].Id)
			case registry.ActionDelete:
				t.Logf("del service:%+v", r.Service.Nodes[0].Id)
			}
		}
		t.Logf("watch quit")
	}()

	time.Sleep(time.Second)
	// add another node
	n1 := &registry.Node{Id: "bbb", Address: "127.0.0.1", Port: 9999}
	s1 := &registry.Service{Name: "test", Nodes: []*registry.Node{n1}}
	r.Register(s1)

	time.Sleep(time.Second)

	// del service
	r.Deregister("aaa")
	r.Deregister("bbb")

	time.Sleep(time.Second)

	// close
	fmt.Printf("close")
	w.Stop()
	r.Stop()
	time.Sleep(time.Second)
}
