package client

import (
	"reflect"

	"github.com/jeckbjy/gsk/util/idgen/xid"

	"github.com/jeckbjy/gsk/arpc/packet"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/selector"
)

func New(opts ...arpc.Option) arpc.Client {
	o := &arpc.Options{}
	for _, fn := range opts {
		fn(o)
	}

	return &_Client{opts: o}
}

type _Client struct {
	opts *arpc.Options
}

func (c *_Client) Init(opts ...arpc.Option) error {
	for _, fn := range opts {
		fn(c.opts)
	}

	return nil
}

// Send 发送消息,不关心结果,也不重试
func (c *_Client) Send(service string, msg interface{}, opts ...arpc.CallOption) error {
	o := arpc.CallOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	next, err := c.getNext(service, &o)
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

// RPC调用,req需要发送的消息,rsp可以是异步回调函数,也可以是变量(同步调用)
func (c *_Client) Call(service string, req interface{}, rsp interface{}, opts ...arpc.CallOption) error {
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

	next, err := c.getNext(service, &o)
	if err != nil {
		return err
	}

	var retryCB arpc.RetryFunc
	if o.Retry > 0 {
		retryCB = func(req arpc.Packet, count int) error {
			// 可以使用backoff方法计算ttl
			req.SetSeqID(xid.New().String())
			return c.sendMsg(next, req)
		}
	}

	// 通过Packet的Internal透传到Filter中注册Callback,可以避免Client依赖Router
	info := &arpc.CallInfo{
		RetryCB:  retryCB,
		Future:   o.Future,
		Response: rsp,
	}

	pkg := c.newRequest(req, &o)
	pkg.SetSeqID(xid.New().String())
	pkg.SetTTL(o.TTL)
	pkg.SetRetry(o.Retry)
	pkg.SetCallInfo(info)

	return c.sendMsg(next, pkg)
}

func (c *_Client) sendMsg(next selector.Next, pkg arpc.Packet) error {
	conn, err := c.getConn(next)
	if err == nil {
		return conn.Send(pkg)
	}

	return err
}

func (c *_Client) getConn(next selector.Next) (anet.Conn, error) {
	node, err := next()
	if err == nil {
		return node.Conn(c.opts.Tran)
	}

	return nil, err
}

func (c *_Client) getNext(service string, opts *arpc.CallOptions) (selector.Next, error) {
	o := c.opts
	if len(o.Proxy) > 0 {
		service = o.Proxy
	}

	return o.Selector.Select(service, &opts.Options)
}

func (c *_Client) newRequest(req interface{}, o *arpc.CallOptions) arpc.Packet {
	if pkg, ok := req.(arpc.Packet); ok {
		if o.ID != 0 {
			pkg.SetMsgID(uint16(o.ID))
		}
		return pkg
	} else {
		// 使用默认的
		pkg := packet.New()
		if o.ID != 0 {
			pkg.SetMsgID(uint16(o.ID))
		} else {
			t := reflect.TypeOf(req)
			if t.Kind() == reflect.Ptr {
				pkg.SetName(t.Elem().Name())
			} else {
				pkg.SetName(t.Name())
			}
		}

		return pkg
	}
}
