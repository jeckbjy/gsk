package nio

import (
	"errors"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/anet/nio/internal"
)

var (
	ErrNoneSelector = errors.New("none selector")
)

func New() anet.Tran {
	t := &nioTran{}
	return t
}

// 可参考: https://github.com/mailru/easygo
// epoll读写方式: https://blog.csdn.net/hzhsan/article/details/23650697
type nioTran struct {
	base.Tran
	loop nioLoop
}

func (t *nioTran) String() string {
	return "nio"
}

func (t *nioTran) Listen(addr string, opts ...anet.ListenOption) (anet.Listener, error) {
	conf := anet.ListenOptions{}
	conf.Init(opts...)

	l, err := internal.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	selector := gLoop.next()
	if selector == nil {
		_ = l.Close()
		return nil, ErrNoneSelector
	}

	return newListener(l, selector, t, conf.Tag)
}

func (t *nioTran) Dial(addr string, opts ...anet.DialOption) (anet.Conn, error) {
	conf := &anet.DialOptions{}
	conf.Init(opts...)
	if conf.Conn == nil {
		poller := gLoop.next()
		conf.Conn = newConn(t, true, conf.Tag, poller)
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
	sock, err := internal.Dial("tcp", addr)
	if err == nil {
		err = conn.Open(sock)
	}

	conf.Call(conn, err)
	return conn, err
}

func (t *nioTran) Close() error {
	return t.Tran.Close()
}
