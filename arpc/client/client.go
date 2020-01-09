package client

import (
	"reflect"
	"time"

	"github.com/jeckbjy/gsk/util/idgen/xid"

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
func (c *_Client) Send(service string, msg interface{}, opts ...arpc.MiscOption) error {
	o := arpc.MiscOptions{}
	o.Init(opts...)

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

// Call - 异步RPC调用
//	req需要发送的消息,可以是arpc.Packet,也可以是普通消息结构体指针
//	rsp可以是异步回调函数,也可以是同步结构体指针
// 	底层必须保证调用了一次Add,则必须对应着调用一次Done,否则会永久等待
// 	Call调用必须有一个超时,防止消息丢失后,永久无法释放
func (c *_Client) Call(service string, req interface{}, rsp interface{}, opts ...arpc.MiscOption) error {
	o := &arpc.MiscOptions{}
	o.Init(opts...)
	o.Response = rsp

	// 同步调用,并且外部没有创建Future
	autoWait := false
	if o.Future == nil && reflect.TypeOf(rsp).Kind() != reflect.Func {
		o.Future = NewFuture()
		autoWait = true
	}

	next, err := c.getNext(service, o)
	if err != nil {
		return err
	}

	if o.RetryNum > 0 {
		o.RetryCB = func(req arpc.Packet) time.Duration {
			// 防止重复接收,使用新的seqID
			req.SetSeqID(xid.New().String())
			if err := c.sendMsg(next, req); err != nil {
				return 0
			}
			// 可以使用backoff方法计算ttl
			return o.TTL
		}
	}

	pkg := c.newRequest(req, o)
	pkg.SetInternal(o)

	err = c.sendMsg(next, pkg)
	if err != nil {
		return err
	}
	if autoWait {
		return o.Future.Wait()
	}
	return nil
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
	} else {
		return nil, err
	}
}

func (c *_Client) getNext(service string, opts *arpc.MiscOptions) (selector.Next, error) {
	o := c.opts
	if len(o.Proxy) > 0 {
		service = o.Proxy
	}

	return o.Selector.Select(service, &opts.Options)
}

func (c *_Client) newRequest(req interface{}, o *arpc.MiscOptions) arpc.Packet {
	if pkg, ok := req.(arpc.Packet); ok {
		return pkg
	} else {
		pkg := arpc.NewPacket()
		// 优先使用MsgID,然后再使用名字
		if o.ID != 0 {
			pkg.SetMsgID(o.ID)
		} else if len(o.Method) != 0 {
			pkg.SetMethod(o.Method)
		} else {
			if t := reflect.TypeOf(req); t.Kind() == reflect.Ptr {
				pkg.SetName(t.Elem().Name())
			} else {
				pkg.SetName(t.Name())
			}
		}
		pkg.SetBody(req)

		return pkg
	}
}
