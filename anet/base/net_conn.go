package base

import (
	"errors"
	"net"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/util/buffer"
)

func NewConn(tran anet.ITran, client bool, tag string) *Conn {
	conn := &Conn{tran: tran, client: client, tag: tag, status: anet.DISCONNECTED}
	conn.rbuf = buffer.New()
	conn.wbuf = buffer.New()
	return conn
}

type Conn struct {
	tran    anet.ITran
	client  bool
	tag     string
	sock    net.Conn
	rbuf    *buffer.Buffer // 读缓存
	wbuf    *buffer.Buffer // 写缓存
	writing bool           // 写线程是否在执行中
	mutex   sync.Mutex     // 锁
	onClose func()         // 关闭时回调,用于自动断线重连
	status  anet.Status
}

func (c *Conn) SetCloseCallback(cb func()) {
	c.onClose = cb
}

func (c *Conn) GetChain() anet.IFilterChain {
	return c.tran.GetChain()
}

func (c *Conn) Status() anet.Status {
	return c.status
}

func (c *Conn) LocalAddr() net.Addr {
	return c.sock.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.sock.RemoteAddr()
}

func (c *Conn) Open(conn net.Conn) {
	if conn == nil {
		return
	}

	c.mutex.Lock()
	c.sock = conn
	c.status = anet.CONNECTED
	c.mutex.Unlock()

	go c.doRead()
	go c.doWrite()

	c.GetChain().HandleOpen(c)
}

func (c *Conn) Close() error {
	c.mutex.Lock()
	sock := c.sock
	c.status = anet.CLOSED
	c.sock = nil
	c.mutex.Unlock()

	if sock != nil {
		return sock.Close()
	} else {
		return errors.New("bad connection when close")
	}
}

func (c *Conn) Read() *buffer.Buffer {
	return c.rbuf
}

func (c *Conn) Write(buf *buffer.Buffer) error {
	c.mutex.Lock()
	c.wbuf.AppendBuffer(buf)
	writing := c.writing
	if !c.writing {
		c.writing = true
	}
	c.mutex.Unlock()

	if !writing {
		go c.doWrite()
	}

	return nil
}

func (c *Conn) Send(msg interface{}) error {
	c.tran.GetChain().HandleWrite(c, msg)
	return nil
}

func (c *Conn) Error(err error) {
	c.GetChain().HandleError(c, err)
}

func (c *Conn) doRead() {
	for {
		// TODO:通过配置分配内存?
		data := make([]byte, 1024)
		n, err := c.sock.Read(data)
		if err != nil {
			c.GetChain().HandleError(c, err)
			c.rbuf.Clear()
			break
		}
		//
		c.rbuf.Append(data[:n])
		c.GetChain().HandleRead(c, c.rbuf)
	}
}

func (c *Conn) doWrite() {
	var err error
	b := buffer.New()
	c.mutex.Lock()
	for {
		if c.wbuf.Empty() {
			break
		}
		buffer.Swap(c.wbuf, b)

		c.mutex.Unlock()
		_, err = b.WriteAll(c.sock)
		c.mutex.Lock()

		if err != nil {
			// 是否需要恢复剩余未发送完成数据?
			break
		}
	}

	c.writing = false
	c.wbuf.Clear()
	c.mutex.Unlock()
	if err != nil {
		c.GetChain().HandleError(c, err)
	}
}
