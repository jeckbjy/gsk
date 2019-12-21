package handler

import (
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
)

var gPool = sync.Pool{New: func() interface{} {
	return &Context{}
}}

func NewContext() arpc.Context {
	return gPool.Get().(arpc.Context)
}

type Context struct {
	conn anet.Conn
	msg  arpc.Packet
}

func (c *Context) Init(conn anet.Conn, msg arpc.Packet) {
	c.conn = conn
	c.msg = msg
}

func (c *Context) Free() {
	gPool.Put(c)
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

func (c *Context) NewPacket() arpc.Packet {
	return arpc.NewPacket()
}
