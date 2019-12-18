package base

import (
	"net"
	"time"

	"github.com/jeckbjy/gsk/anet"
)

type ListenCB func() (net.Listener, error)
type DialCB func() (net.Conn, error)

// base transport
type Tran struct {
	chain     anet.FilterChain
	listeners []net.Listener
}

func (t *Tran) GetChain() anet.FilterChain {
	return t.chain
}

func (t *Tran) SetChain(chain anet.FilterChain) {
	if chain != nil {
		t.chain = chain
	}
}

func (t *Tran) AddFilters(filters ...anet.Filter) {
	if t.chain == nil {
		t.chain = NewFilterChain()
	}

	t.chain.AddLast(filters...)
}

func (t *Tran) Close() error {
	var err error
	for _, l := range t.listeners {
		if e := l.Close(); e != nil {
			err = e
		}
	}

	t.listeners = nil

	return err
}

func (t *Tran) AddListener(l net.Listener) {
	t.listeners = append(t.listeners, l)
}

func DialTCP(addr string, timeout time.Duration) (net.Conn, error) {
	if timeout != 0 {
		return net.DialTimeout("tcp", addr, timeout)
	} else {
		return net.Dial("tcp", addr)
	}
}
