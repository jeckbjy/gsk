package router

import (
	"fmt"
	"math"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/util/container/arrmap"
)

const abortIndex int8 = math.MaxInt8 / 2

var gContextPool = sync.Pool{
	New: func() interface{} {
		return &Context{index: -1}
	},
}

func NewContext() arpc.Context {
	return gContextPool.Get().(arpc.Context)
}

type Context struct {
	arrmap.StringMap
	handler arpc.HandlerFunc
	chain   arpc.HandlerChain
	conn    anet.Conn
	msg     arpc.Packet
	rsp     arpc.Packet
	data    interface{}
	err     error
	index   int8
}

func (c *Context) Init(conn anet.Conn, msg arpc.Packet) {
	c.conn = conn
	c.msg = msg
	c.handler = nil
	c.chain = nil
	c.rsp = nil
	c.data = nil
	c.err = nil
	c.index = -1
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

func (c *Context) Response() arpc.Packet {
	return c.rsp
}

func (c *Context) SetResponse(rsp arpc.Packet) {
	c.rsp = rsp
}

func (c *Context) Send(msg interface{}) error {
	return c.conn.Send(msg)
}

func (c *Context) Handler() arpc.HandlerFunc {
	return c.handler
}

func (c *Context) SetHandler(h arpc.HandlerFunc) {
	c.handler = h
}

func (c *Context) SetMiddleware(hc arpc.HandlerChain) {
	c.chain = hc
}

func (c *Context) Abort(err error) {
	c.err = err
	c.index = abortIndex
}

func (c *Context) Next() error {
	c.index++
	size := int8(len(c.chain))
	for s := size + 1; c.index < s; c.index++ {
		if c.index == size {
			// 执行handler
			if c.handler == nil {
				err := fmt.Errorf(
					"no handler,ack=%v,msgid=%d,name=%s,seqid=%v",
					c.msg.IsAck(), c.msg.MsgID(), c.msg.Name(), c.msg.SeqID(),
				)
				c.err = err
				return err
			}
			return c.handler(c)
		} else {
			if err := c.chain[c.index](c); err != nil {
				c.Abort(err)
			}
		}
	}

	return nil
}
