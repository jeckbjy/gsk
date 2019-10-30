package nio

import (
	"net"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/anet/nio/internal"
)

func init() {
	anet.Add("nio", New)
}

func New() anet.Tran {
	return &nTran{}
}

// TODO:目前还有很多问题,以后再修正
// 可参考: https://github.com/mailru/easygo
type nTran struct {
	base.Tran
	selector []*internal.Selector
	index    int
	mux      sync.Mutex
}

func (t *nTran) String() string {
	return "tcp"
}

func (t *nTran) Listen(addr string, opts ...anet.ListenOption) (anet.Listener, error) {
	conf := anet.ListenOptions{}
	conf.Init(opts...)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			s := t.next()
			c := &nConn{selector: s}
			c.Open(conn)
			s.Add(conn, internal.OP_RW, c)
		}
	}()

	return nil, nil
}

func (t *nTran) Dial(addr string, opts ...anet.DialOption) (anet.Conn, error) {
	conf := &anet.DialOptions{}
	conf.Init(opts...)
	return nil, nil
}

func (t *nTran) Close() error {
	return nil
}

func (t *nTran) next() *internal.Selector {
	t.mux.Lock()
	defer t.mux.Unlock()
	s := t.selector[t.index]
	t.index++
	return s
}
