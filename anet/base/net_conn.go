package base

import (
	"net"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/util/buffer"
)

func NewConn(tran anet.Tran, client bool, tag string) *Conn {
	conn := &Conn{tran: tran, client: client, tag: tag, status: anet.CONNECTING}
	conn.cond = sync.NewCond(&conn.mutex)
	conn.rbuf = buffer.New()
	conn.wbuf = buffer.New()
	return conn
}

// 基于标准net.Conn的实现,由于是阻塞IO,读写各自起了一个协程
type Conn struct {
	tran   anet.Tran              // transport
	sock   net.Conn               // 原始socket
	rbuf   *buffer.Buffer         // 读缓存
	wbuf   *buffer.Buffer         // 写缓存
	mutex  sync.Mutex             // 锁
	cond   *sync.Cond             // 用于通知写协程退出
	status anet.Status            // 当前状态
	client bool                   // 用于标识是服务器端接收的连接还是客户端Dial产生的链接
	tag    string                 // 用于给外部标识类型
	data   map[string]interface{} // 自定义数据
}

func (c *Conn) GetChain() anet.FilterChain {
	return c.tran.GetChain()
}

func (c *Conn) Tran() anet.Tran {
	return c.tran
}

func (c *Conn) Tag() string {
	return c.tag
}

func (c *Conn) Get(key string) interface{} {
	if c.data != nil {
		return c.data[key]
	}

	return nil
}

func (c *Conn) Set(key string, val interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	c.data[key] = val
}

func (c *Conn) Status() anet.Status {
	c.mutex.Lock()
	status := c.status
	c.mutex.Unlock()
	return status
}

func (c *Conn) IsActive() bool {
	return c.Status() == anet.OPEN
}

func (c *Conn) IsDial() bool {
	return c.client
}

func (c *Conn) LocalAddr() net.Addr {
	return c.sock.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.sock.RemoteAddr()
}

func (c *Conn) Open(conn net.Conn) error {
	var err error
	c.mutex.Lock()
	if c.sock == nil {
		c.sock = conn
		c.status = anet.OPEN
		c.rbuf.Clear()
	} else {
		err = anet.ErrHasOpened
	}

	c.mutex.Unlock()

	if err == nil {
		go c.doRead(conn)
		go c.doWrite()
		c.GetChain().HandleOpen(c)
	}

	return err
}

func (c *Conn) Close() error {
	c.mutex.Lock()

	status := c.status
	if c.status == anet.OPEN {
		c.status = anet.CLOSING
	}

	c.mutex.Unlock()
	// 通知写线程退出
	// TODO:阻塞等待?
	if status == anet.OPEN {
		c.cond.Signal()
	}

	return nil
}

func (c *Conn) Read() *buffer.Buffer {
	return c.rbuf
}

func (c *Conn) Write(buf *buffer.Buffer) error {
	var err error
	c.mutex.Lock()
	// 连接过程中也可以发送,等连接成功后会主动发送所有数据
	// 如果连接失败则会清空数据
	if c.status == anet.CONNECTING || c.status == anet.OPEN {
		c.wbuf.AppendBuffer(buf)
	} else {
		err = anet.ErrHasClosed
	}

	c.mutex.Unlock()

	if err == nil {
		c.cond.Signal()
	}

	return err
}

func (c *Conn) Send(msg interface{}) error {
	c.tran.GetChain().HandleWrite(c, msg)
	return nil
}

func (c *Conn) Clear() {
	c.mutex.Lock()
	c.wbuf.Clear()
	c.rbuf.Clear()
	c.mutex.Unlock()
}

func (c *Conn) Error(err error) {
	// 是否需要清空写buf?
	//c.mutex.Lock()
	//if c.status == anet.CONNECTING {
	//	c.wbuf.Clear()
	//}
	//c.mutex.Unlock()
	c.GetChain().HandleError(c, err)
}

func (c *Conn) doClose(err error) anet.Status {
	c.mutex.Lock()
	status := c.status
	if c.status == anet.CLOSED {
		c.mutex.Unlock()
		return status
	}

	c.wbuf.Clear()
	c.rbuf.Clear()
	c.status = anet.CLOSED
	if c.sock != nil {
		_ = c.sock.Close()
		c.sock = nil
	}
	c.mutex.Unlock()

	if err != nil {
		c.Error(err)
	}

	return status
}

func (c *Conn) doRead(sock net.Conn) {
	// TODO:通过配置分配内存?
	chunk := 1024
	for {
		data := make([]byte, chunk)
		n, err := sock.Read(data)

		if err != nil {
			status := c.doClose(err)
			if status == anet.OPEN {
				// 通知write退出
				c.cond.Signal()
			}
			break
		}

		if c.Status() != anet.OPEN {
			// 丢弃closing状态中发来的消息,并退出
			c.Error(anet.ErrHasClosed)
			break
		}

		c.rbuf.Append(data[:n])
		c.GetChain().HandleRead(c, c.rbuf)
	}
}

func (c *Conn) doWrite() {
	b := buffer.New()
	for {
		c.mutex.Lock()
		for (c.status == anet.CONNECTING) || (c.status == anet.OPEN && c.wbuf.Empty()) {
			c.cond.Wait()
		}

		buffer.Swap(c.wbuf, b)
		sock := c.sock
		status := c.status
		c.mutex.Unlock()

		// 发送剩余数据
		if sock != nil && !b.Empty() {
			_, err := b.WriteAll(sock)
			if err != nil {
				c.doClose(err)
				break
			}
		}

		if status == anet.CLOSING {
			c.doClose(nil)
			break
		}

		b.Clear()
	}
}
