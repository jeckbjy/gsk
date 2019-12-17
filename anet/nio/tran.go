package nio

import (
	"log"
	"net"
	"runtime"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/anet/nio/internal"
)

func New() anet.Tran {
	return &nioTran{max: runtime.NumCPU(), connMap: make(map[uintptr]*nioConn)}
}

// 可参考: https://github.com/mailru/easygo
// 正确的epoll读写方式: https://blog.csdn.net/hzhsan/article/details/23650697
type nioTran struct {
	base.Tran
	selector []*internal.Selector
	index    int
	max      int
	mux      sync.Mutex
	connMap  map[uintptr]*nioConn
	connMux  sync.Mutex
}

func (t *nioTran) String() string {
	return "nio"
}

func (t *nioTran) newConn(client bool, tag string) *nioConn {
	s := t.next()
	if s == nil {
		return nil
	}
	return newConn(t, client, tag, s)
}

func (t *nioTran) Listen(addr string, opts ...anet.ListenOption) (anet.Listener, error) {
	conf := anet.ListenOptions{}
	conf.Init(opts...)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			sock, err := l.Accept()
			if err != nil {
				log.Print(err)
				return
			}

			conn := t.newConn(false, conf.Tag)
			if conn == nil {
				log.Printf("create selector fail")
				continue
			}
			if err = conn.Open(sock); err != nil {
				log.Print(err)
				continue
			}
			t.addConn(conn)
			if err := conn.selector.Add(conn.fd); err != nil {
				log.Print(err)
				t.delConn(conn.fd)
				conn.onError(err)
			}
		}
	}()

	return nil, nil
}

func (t *nioTran) Dial(addr string, opts ...anet.DialOption) (anet.Conn, error) {
	conf := &anet.DialOptions{}
	conf.Init(opts...)
	if conf.Conn == nil {
		conf.Conn = t.newConn(true, conf.Tag)
	}

	if conf.Blocking {
		return t.doDial(conf, addr)
	} else {
		go t.doDial(conf, addr)
		return conf.Conn, nil
	}
}

func (t *nioTran) doDial(conf *anet.DialOptions, addr string) (anet.Conn, error) {
	conn := conf.Conn.(*nioConn)
	sock, err := base.DialTCP(addr, conf.Timeout)
	if err == nil {
		err = conn.Open(sock)
	}

	t.addConn(conn)
	if err := conn.selector.Add(conn.fd); err != nil {
		t.delConn(conn.fd)
		conn.onError(err)
	}
	return conn, nil
}

func (t *nioTran) Close() error {
	t.connMux.Lock()
	for fd, conn := range t.connMap {
		_ = conn.selector.Delete(fd)
	}
	t.connMap = make(map[uintptr]*nioConn)
	t.connMux.Unlock()
	return t.Tran.Close()
}

func (t *nioTran) addConn(conn *nioConn) {
	t.connMux.Lock()
	t.connMap[conn.fd] = conn
	t.connMux.Unlock()
}

func (t *nioTran) delConn(fd uintptr) {
	t.connMux.Lock()
	delete(t.connMap, fd)
	t.connMux.Unlock()
}

func (t *nioTran) getConn(fd uintptr) *nioConn {
	t.connMux.Lock()
	conn := t.connMap[fd]
	t.connMux.Unlock()
	return conn
}

func (t *nioTran) next() *internal.Selector {
	t.mux.Lock()

	if len(t.selector) == 0 {
		t.selector = make([]*internal.Selector, t.max)
	}

	index := t.index % len(t.selector)
	s := t.selector[index]
	t.index++
	if s == nil {
		selector, err := internal.New()
		if err == nil {
			t.selector[index] = selector
			s = selector
			go t.loop(selector)
		} else {
			log.Print(err)
		}
	}
	t.mux.Unlock()
	return s
}

func (t *nioTran) loop(s *internal.Selector) {
	for {
		err := s.Wait(func(event *internal.Event) {
			conn := t.getConn(event.Fd())
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
