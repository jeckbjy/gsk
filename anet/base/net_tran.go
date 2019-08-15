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
	chain     anet.IFilterChain
	listeners []net.Listener
}

func (t *Tran) GetChain() anet.IFilterChain {
	return t.chain
}

func (t *Tran) SetChain(chain anet.IFilterChain) {
	if chain != nil {
		t.chain = chain
	}
}

func (t *Tran) AddFilters(filters ...anet.IFilter) {
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

func DoOpen(c net.Conn, tran anet.ITran, client bool, tag string) *Conn {
	conn := NewConn(tran, client, tag)
	conn.Open(c)
	return conn
}

// DoListen auxiliary function for listen
func DoListen(conf *anet.ListenOptions, t anet.ITran, cb ListenCB) (anet.IListener, error) {
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
func DoDial(conf *anet.DialOptions, t anet.ITran, cb DialCB) (anet.IConn, error) {

	if conf.Blocking {
		var conn anet.IConn
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
		}

		if conn == nil {
			conn = NewConn(t, true, conf.Tag)
		}

		if conf.RetryMax != 0 {
			conn.SetCloseCallback(func() {
				_, _ = DoAsyncDial(conn, conf, cb)
			})
		}

		return DoAsyncDial(conn, conf, cb)
	}
}

// DoAsyncDial 尝试异步连接
func DoAsyncDial(conn *Conn, conf *anet.DialOptions, cb DialCB) (anet.IConn, error) {
	go func() {
		count := 0
		for {
			count++
			c, err := cb()
			if err == nil {
				// dial success
				conn.Open(c)
				if conf.Callback != nil {
					conf.Callback(conn, nil)
				}
				break
			}

			if conf.Callback != nil {
				conf.Callback(conn, err)
			}

			// handle error
			conn.Error(err)
			if count > conf.RetryMax {
				break
			}

			// wait for next
			time.Sleep(conf.Interval)
		}
	}()

	return conn, nil
}
