package base

import (
	"errors"
	"net"

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

	return err
}

func DoOpen(c net.Conn, tran anet.Tran, client bool, tag string) *Conn {
	conn := NewConn(tran, client, tag)
	conn.Open(c)
	return conn
}

// DoListen auxiliary function for listen
func DoListen(conf *anet.ListenOptions, t anet.Tran, cb ListenCB) (anet.Listener, error) {
	l, err := cb()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			DoOpen(c, t, false, conf.Tag)
		}
	}()

	return l, nil
}

// DoDial auxiliary function for dial
func DoDial(conf *anet.DialOptions, t anet.Tran, cb DialCB) (anet.Conn, error) {
	if conf.Blocking {
		var conn anet.Conn
		c, err := cb()
		if err == nil {
			conn = DoOpen(c, t, true, conf.Tag)
		}

		if conf.Callback != nil {
			conf.Callback(conn, err)
		}
		return conn, nil
	} else {
		var conn *Conn
		// 使用老的Conn
		if conf.Conn != nil {
			conn = conf.Conn.(*Conn)
			if !conn.IsDial() {
				return conn, errors.New("bad dialer")
			}
		}

		if conn == nil {
			conn = NewConn(t, true, conf.Tag)
		}

		return DoAsyncDial(conn, conf, cb)
	}
}

// DoAsyncDial 尝试异步连接
func DoAsyncDial(conn *Conn, conf *anet.DialOptions, cb DialCB) (anet.Conn, error) {
	go func() {
		c, err := cb()
		if err == nil {
			// dial success
			if err := conn.Open(c); err != nil {
				if conf.Callback != nil {
					conf.Callback(conn, nil)
				}
			}
		} else {
			if conf.Callback != nil {
				conf.Callback(conn, err)
			}

			conn.Error(err)
		}
	}()

	return conn, nil
}
