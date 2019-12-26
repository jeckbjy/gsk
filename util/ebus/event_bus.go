package ebus

import (
	"container/list"
	"sync"
)

func New() IEventBus {
	eb := &eventBus{handlers: make(map[string]*list.List)}
	return eb
}

type eventHandler struct {
	id       uint
	once     bool
	priority int
	cb       Callback
}

type eventBus struct {
	mux      sync.Mutex
	maxID    uint
	handlers map[string]*list.List
}

func (eb *eventBus) Len(topic string) int {
	eb.mux.Lock()
	ls, ok := eb.handlers[topic]
	num := 0
	if ok {
		num = ls.Len()
	}
	eb.mux.Unlock()
	return num
}

func (eb *eventBus) Listen(topic string, once bool, priority int, cb Callback) uint {
	if cb == nil || topic == "" {
		return 0
	}

	eb.mux.Lock()
	eb.maxID++
	eh := &eventHandler{id: eb.maxID, once: once, cb: cb}
	ls, ok := eb.handlers[topic]
	if !ok {
		ls = list.New()
		eb.handlers[topic] = ls
	}

	// 从后向前查找
	inserted := false
	for iter := ls.Back(); iter != nil; iter = iter.Prev() {
		if iter.Value.(*eventHandler).priority <= priority {
			ls.InsertAfter(eh, iter)
			inserted = true
			break
		}
	}
	if !inserted {
		ls.PushFront(eh)
	}
	eb.mux.Unlock()
	return eh.id
}

func (eb *eventBus) Remove(topic string, id uint) {
	eb.mux.Lock()
	handlers, ok := eb.handlers[topic]
	if ok && handlers.Len() > 0 {
		for iter := handlers.Front(); iter != nil; iter = iter.Next() {
			if iter.Value.(*eventHandler).id == id {
				handlers.Remove(iter)
				break
			}
		}
		if handlers.Len() == 0 {
			delete(eb.handlers, topic)
		}
	}
	eb.mux.Unlock()
}

func (eb *eventBus) Clear(topic string) {
	eb.mux.Lock()
	delete(eb.handlers, topic)
	eb.mux.Unlock()
}

func (eb *eventBus) Emit(topic string, ev IEvent) {
	eb.mux.Lock()
	handlers, ok := eb.handlers[topic]
	if ok && handlers.Len() > 0 {
		for iter := handlers.Front(); iter != nil; {
			curr := iter
			iter = iter.Next()

			eh := curr.Value.(*eventHandler)
			eh.cb(ev)
			if eh.once {
				handlers.Remove(curr)
			}
		}

		if handlers.Len() == 0 {
			delete(eb.handlers, topic)
		}
	}
	eb.mux.Unlock()
}
