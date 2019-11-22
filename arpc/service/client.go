package service

import (
	"reflect"
	"time"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/selector"
	"github.com/jeckbjy/gsk/util/idgen/xid"
)

type Client struct {
	opts *arpc.Options
}

// Send 发送消息,不关心结果,也不重试
func (c *Client) Send(service string, msg interface{}, opts ...arpc.CallOption) error {
	o := arpc.CallOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	next, err := c.getNext(service)
	if err != nil {
		return err
	}

	pkg := c.newRequest(msg, &o)
	conn, err := c.getConn(next)
	if err != nil {
		return err
	}
	return conn.Send(pkg)
}

func (c *Client) Call(service string, req interface{}, rsp interface{}, opts ...arpc.CallOption) error {
	o := arpc.CallOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	if o.TTL == 0 {
		o.TTL = arpc.DefaultCallTTL
	}

	if o.Future == nil && reflect.TypeOf(rsp).Kind() != reflect.Func {
		o.Future = NewFuture()
	}

	next, err := c.getNext(service)
	if err != nil {
		return err
	}
	pkg := c.newRequest(req, &o)
	pkg.SetSeqID(xid.New().String())

	// 支持重发?
	var retry arpc.RetryFunc
	if o.Retry > 0 {
		retry = func(req arpc.Packet, count int) (time.Duration, error) {
			// TODO:backoff??
			ttl := o.TTL
			req.SetSeqID(xid.New().String())
			err := c.sendMsg(next, req)
			return ttl, err
		}
	}

	ropts := arpc.RegisterOptions{
		SeqID:    pkg.SeqID(),
		TTL:      o.TTL,
		Future:   o.Future,
		RetryMax: o.Retry,
		RetryCB:  retry,
	}

	if err := c.opts.Router.Register(rsp, &ropts); err != nil {
		return err
	}

	return c.sendMsg(next, pkg)
}

func (c *Client) sendMsg(next selector.Next, pkg arpc.Packet) error {
	conn, err := c.getConn(next)
	if err != nil {
		return err
	}
	return conn.Send(pkg)
}

func (c *Client) getConn(next selector.Next) (anet.Conn, error) {
	//node, err := next()
	//if err != nil {
	//	return nil, err
	//}
	//
	//if node.Conn() == nil {
	//	if conn, err := c.opts.Tran.Dial(node.Address); err != nil {
	//		return nil, err
	//	} else {
	//		node.SetConn(conn)
	//	}
	//}
	//
	//return node.Conn().(anet.Conn), nil
}

func (c *Client) getNext(service string) (selector.Next, error) {
	o := c.opts
	if len(o.Proxy) > 0 {
		service = o.Proxy
	}

	return o.Selector.Select(service)
}

func (c *Client) newRequest(req interface{}, o *arpc.CallOptions) arpc.Packet {
	var pkg arpc.Packet
	if p, ok := req.(arpc.Packet); ok {
		pkg = p
	} else if c.opts.PacketFunc != nil {
		pkg = c.opts.PacketFunc()
	} else {
		return nil
	}

	if o.ID != 0 {
		pkg.SetMsgID(uint16(o.ID))
	}

	if o.Name != "" {
		pkg.SetName(o.Name)
	}

	if o.Method != "" {
		pkg.SetMethod(o.Method)
	}

	return pkg
}
