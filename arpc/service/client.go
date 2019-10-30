package service

import (
	"reflect"
	"time"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/selector"
	"github.com/jeckbjy/gsk/util/id"
)

type Client struct {
	opts *arpc.Options
}

// Send 发送消息,不关心结果,也不重试
func (c *Client) Send(service string, msg interface{}) error {
	next, err := c.getNext(service)
	if err != nil {
		return err
	}

	pkg := c.newRequest(msg, nil)
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

	if o.Future == nil && reflect.TypeOf(rsp).Kind() != reflect.Func {
		o.Future = NewFuture()
	}

	next, err := c.getNext(service)
	if err != nil {
		return err
	}

	pkg := c.newRequest(req, &o)
	pkg.SetSeqID(id.NewXID())

	// 支持重发?
	var retry arpc.RetryFunc
	if o.Retry > 0 {
		retry = func(count int) (string, time.Duration, error) {
			seqId := pkg.SeqID()
			ttl := o.TTL // TODO:backoff??
			pkg.SetSeqID(id.NewXID())
			err := c.sendMsg(next, pkg)
			return seqId, ttl, err
		}
	}

	ops := arpc.RegisterRPCOptions{
		SeqID:    pkg.SeqID(),
		TTL:      o.TTL,
		Future:   o.Future,
		RetryMax: o.Retry,
		RetryCB:  retry,
	}

	if err := c.opts.RPCRouter.Register(rsp, &ops); err != nil {
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
	node, err := next()
	if err != nil {
		return nil, err
	}

	if node.Conn() == nil {
		if conn, err := c.opts.Tran.Dial(node.Address); err != nil {
			return nil, err
		} else {
			node.SetConn(conn)
		}
	}

	return node.Conn().(anet.Conn), nil
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
	} else {
		// create
	}

	return pkg
}
