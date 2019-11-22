package local

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/jeckbjy/gsk/registry"

	"github.com/jeckbjy/gsk/util/ssdp"
)

func TestRegister(t *testing.T) {
	ssdp.Logger = log.New(os.Stderr, "[SSDP] ", log.LstdFlags)
	t.Log("start register")
	r := New(nil)
	err := r.Register(registry.NewService("test", "aaa", "127.0.0.1:9999", nil))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("start watch")
	if err := r.Watch([]string{"test"}, func(ev *registry.Event) {
		t.Logf("new event,%+v,%+v,%+v", ev.Type, ev.Id, ev.Service)
	}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)

	// add another
	_ = r.Register(registry.NewService("test", "bbb", "127.0.0.1:9999", nil))
	time.Sleep(time.Second)

	_ = r.Unregister("aaa")
	_ = r.Unregister("bbb")
	time.Sleep(time.Second)

	_ = r.Close()
	time.Sleep(time.Second)
}
