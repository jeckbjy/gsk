package nio

import (
	"log"
	"runtime"
	"sync"

	"github.com/jeckbjy/gsk/anet/nio/internal"
)

// 使用全局的EventLoop
var gLoop = newLoop()

func newLoop() *nioLoop {
	l := &nioLoop{}
	l.init()
	return l
}

// 只能在初始化前设置一次
func SetMaxLoopNum(num int) {
	gLoop.setMax(num)
}

type nioChannel interface {
	onEvent(ev *internal.Event)
}

type nioLoop struct {
	mux      sync.Mutex
	channels map[internal.FD]nioChannel
	pollers  []internal.Poller
	index    int
	max      int
}

func (l *nioLoop) init() {
	l.max = runtime.NumCPU()
	l.channels = make(map[internal.FD]nioChannel)
}

func (l *nioLoop) setMax(max int) {
	l.max = max
}

func (l *nioLoop) add(fd internal.FD, channel nioChannel) {
	l.mux.Lock()
	l.channels[fd] = channel
	l.mux.Unlock()
}

func (l *nioLoop) remove(fd internal.FD) {
	l.mux.Lock()
	delete(l.channels, fd)
	l.mux.Unlock()
}

func (l *nioLoop) get(fd internal.FD) nioChannel {
	l.mux.Lock()
	channel := l.channels[fd]
	l.mux.Unlock()
	return channel
}

func (l *nioLoop) next() internal.Poller {
	l.mux.Lock()
	if len(l.pollers) == 0 {
		l.pollers = make([]internal.Poller, l.max)
	}
	index := l.index % len(l.pollers)
	s := l.pollers[index]
	l.index++
	if s == nil {
		selector, err := internal.New()
		if err == nil {
			l.pollers[index] = selector
			s = selector
			go l.run(selector)
		} else {
			log.Print(err)
		}
	}
	l.mux.Unlock()
	return s
}

func (l *nioLoop) run(s internal.Poller) {
	for {
		err := s.Wait(func(event *internal.Event) {
			conn := l.get(event.Fd())
			if conn != nil {
				conn.onEvent(event)
			} else {
				log.Printf("not found conn,%+v", event.Fd())
				_ = event.Delete()
			}
		})

		if err != nil {
			log.Print(err)
		}
	}
}
