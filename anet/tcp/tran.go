package tcp

import (
	"net"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
)

func New() anet.Tran {
	return &Tran{}
}

// Tran tcp Transport
type Tran struct {
	base.Tran
}

func (t *Tran) String() string {
	return "tcp"
}

func (t *Tran) NewConn(client bool, tag string) anet.Conn {
	return base.NewNetConn(t, client, tag)
}

func (t *Tran) Listen(addr string, opts ...anet.ListenOption) (anet.Listener, error) {
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
				return
			}

			conn := base.NewNetConn(t, false, conf.Tag)
			_ = conn.Open(sock)
		}
	}()

	return l, nil
}

func (t *Tran) Dial(addr string, opts ...anet.DialOption) (anet.Conn, error) {
	conf := &anet.DialOptions{}
	conf.Init(opts...)

	if conf.Conn == nil {
		conf.Conn = base.NewNetConn(t, true, conf.Tag)
	}

	if conf.Blocking {
		return t.doDial(conf, addr)
	} else {
		go t.doDial(conf, addr)
		return conf.Conn, nil
	}
}

func (t *Tran) doDial(conf *anet.DialOptions, addr string) (anet.Conn, error) {
	conn := conf.Conn.(*base.NetConn)
	sock, err := base.DialTCP(addr, conf.Timeout)
	if err == nil {
		err = conn.Open(sock)
	}

	conf.Call(conn, nil)
	return conn, nil
}
