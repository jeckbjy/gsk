package ebus

import (
	"log"
	"testing"
)

func TestEventBus(t *testing.T) {
	id := Listen(onEvent)
	t.Log("id", id)
	Emit(&TestEvent{Data: "test"})
	Remove("TestEvent", id)
	t.Log("len", Len("TestEvent"))
}

type TestEvent struct {
	Data string
}

func onEvent(ev *TestEvent) {
	log.Printf("onEvent, %+v", ev.Data)
}
