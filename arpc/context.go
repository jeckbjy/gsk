package arpc

import (
	"github.com/jeckbjy/gsk/anet"
)

func NewContext(conn anet.IConn, req IPacket, rsp IPacket) *Context {
	return &Context{conn, req, rsp, nil}
}

type Context struct {
	conn    anet.IConn
	req     IPacket
	rsp     IPacket
	handler Handler
}

func (c *Context) Reset() {
	c.conn = nil
	c.req = nil
	c.rsp = nil
	c.handler = nil
}

func (c *Context) Conn() anet.IConn {
	return c.conn
}

func (c *Context) Request() IPacket {
	return c.req
}

func (c *Context) Response() IPacket {
	return c.rsp
}

func (c *Context) Send(msg interface{}) error {
	return c.conn.Send(msg)
}

func (c *Context) Handler() Handler {
	return c.handler
}

func (c *Context) SetHandler(h Handler) {
	c.handler = h
}
