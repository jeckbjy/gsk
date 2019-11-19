package handler

import (
	"context"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
)

func NewContext(conn anet.Conn, req arpc.Packet, rsp arpc.Packet) arpc.Context {
	return &Context{Context: context.Background(), conn: conn, req: req, rsp: rsp}
}

type Context struct {
	context.Context
	conn anet.Conn
	req  arpc.Packet
	rsp  arpc.Packet
}

func (c *Context) Reset() {
	c.conn = nil
}

func (c *Context) Conn() anet.Conn {
	return c.conn
}

func (c *Context) Request() arpc.Packet {
	return c.req
}

func (c *Context) Response() arpc.Packet {
	return c.rsp
}

func (c *Context) Send(msg interface{}) error {
	return c.conn.Send(msg)
}
