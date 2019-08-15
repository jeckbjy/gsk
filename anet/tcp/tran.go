package tcp

import (
	"net"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
)

func init() {
	anet.Add("tcp", New)
}

func New() anet.ITran {
	return &Tran{}
}

// Tran tcp Transport
type Tran struct {
	base.Tran
}

func (t *Tran) String() string {
	return "tcp"
}

func (t *Tran) Listen(addr string, opts ...anet.ListenOption) (anet.IListener, error) {
	conf := anet.ListenOptions{}
	conf.Init(opts...)
	return base.DoListen(&conf, t, func() (net.Listener, error) {
		return net.Listen("tcp", addr)
	})
}

func (t *Tran) Dial(addr string, opts ...anet.DialOption) (anet.IConn, error) {
	conf := &anet.DialOptions{}
	conf.Init(opts...)
	return base.DoDial(conf, t, func() (net.Conn, error) {
		if conf.Timeout != 0 {
			return net.DialTimeout("tcp", addr, conf.Timeout)
		} else {
			return net.Dial("tcp", addr)
		}
	})
}
