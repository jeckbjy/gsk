package router

import (
	"sync"

	"github.com/jeckbjy/gsk/util/container/arrmap"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
)

var gContextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

func NewContext() arpc.Context {
	return gContextPool.Get().(arpc.Context)
}

type Context struct {
	arrmap.StringMap
	conn anet.Conn
	msg  arpc.Packet
	data interface{}
	err  error
}

func (c *Context) Init(conn anet.Conn, msg arpc.Packet) {
	c.conn = conn
	c.msg = msg
	c.data = nil
	c.err = nil
}

func (c *Context) Free() {
	gContextPool.Put(c)
}

func (c *Context) Data() interface{} {
	return c.data
}

func (c *Context) SetData(v interface{}) {
	c.data = v
}

func (c *Context) Error() error {
	return c.err
}

func (c *Context) SetError(err error) {
	c.err = err
}

func (c *Context) Conn() anet.Conn {
	return c.conn
}

func (c *Context) Message() arpc.Packet {
	return c.msg
}

func (c *Context) Send(msg interface{}) error {
	return c.conn.Send(msg)
}
