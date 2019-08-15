package arpc

import (
	"reflect"
	"time"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/arpc/selector"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/id"
)

func NewClient(opts ...ClientOption) IClient {
	c := &Client{opts: &ClientOptions{}}
	c.Init(opts...)
	return c
}

// Client 需要实现服务发现,消息事件监听,需要监听哪些服务最好启动时注册，
type Client struct {
	opts *ClientOptions
}

func (c *Client) Options() ClientOptions {
	return *c.opts
}

func (c *Client) Init(opts ...ClientOption) {
	c.opts.Init(opts...)
}

// Send 发送一条消息,不关心返回结果
func (c *Client) Send(service string, msg interface{}, opts ...MiscOption) error {
	conf := MiscOptions{}
	conf.Init(opts...)
	if conf.RetryMax < 1 {
		conf.RetryMax = 1
	}

	next, err := c.Select(service)
	if err != nil {
		return err
	}

	pkg := c.BuildRequest(msg, &conf)

	for i := 0; i < conf.RetryMax; i++ {
		node, err := next()
		if err == selector.ErrNoneAvailable {
			break
		} else if err != nil {
			continue
		}

		conn := c.GetConn(node)
		if conn != nil {
			_ = conn.Send(pkg)
		}
	}

	return nil
}

// Call RPC调用
func (c *Client) Call(service string, req interface{}, rsp interface{}, opts ...MiscOption) error {
	conf := MiscOptions{}
	conf.Init(opts...)
	if conf.RetryMax < 1 {
		conf.RetryMax = 1
	}
	if conf.Err == nil {
		conf.Err = ErrTimeout
	}

	next, err := c.Select(service)
	if err != nil {
		return err
	}

	pkg := c.BuildRequest(req, &conf)
	pkg.SetSeqID(id.NewXID())

	count := 0
	check := func() (string, time.Duration, error) {
		if count < conf.RetryMax {
			_ = conf.Future.Fail(conf.Err)
			return "", 0, ErrRetryStop
		}
		count++
		return id.NewXID(), conf.TTL, nil
	}

	retry := func() error {
		node, err := next()
		if err != nil {
			_ = conf.Future.Fail(conf.Err)
			return ErrNoNode
		}

		conn := c.GetConn(node)
		if conn == nil {
			_ = conf.Future.Fail(conf.Err)
			return ErrNoConn
		}
		return conn.Send(pkg)
	}

	rops := RegisterRpcOptions{
		SeqID:  pkg.SeqID(),
		TTL:    conf.TTL,
		Future: conf.Future,
		Retry:  retry,
		Check:  check,
	}

	if err := c.opts.Router.RegisterRpc(rsp, &rops); err != nil {
		return err
	}

	return retry()
}

// 如果不存在,则创建连接,每个node一个连接?
func (c *Client) GetConn(node *registry.Node) anet.IConn {
	if node.Conn() == nil {
		if conn, err := c.opts.Tran.Dial(node.Address); err != nil {
			return nil
		} else {
			node.SetConn(conn)
		}
	}

	return node.Conn().(anet.IConn)
}

func (c *Client) BuildRequest(req interface{}, op *MiscOptions) IPacket {
	var pkg IPacket
	if p, ok := req.(IPacket); ok {
		pkg = p
	} else {
		pkg = c.opts.Creator()
		if op.Name != "" {
			op.Name = reflect.TypeOf(req).Name()
		}

		pkg.SetBody(req)
	}

	// auto find msgid?

	if op.ID != 0 {
		pkg.SetID(op.ID)
	}

	if op.Name != "" {
		pkg.SetName(op.Name)
	}

	if op.Method != "" {
		pkg.SetMethod(op.Method)
	}

	return pkg
}

func (c *Client) Select(service string) (selector.Next, error) {
	if len(c.opts.Proxy) > 0 {
		return c.opts.Selector.Select(c.opts.Proxy)
	} else {
		return c.opts.Selector.Select(service)
	}
}
