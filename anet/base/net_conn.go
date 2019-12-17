package base

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/util/buffer"
)

func NewNetConn(tran anet.Tran, client bool, tag string) *NetConn {
	conn := &NetConn{}
	conn.Init(tran, client, tag)
	return conn
}

// 封装基础功能
type Conn struct {
	tran       anet.Tran              // transport
	rmux       sync.Mutex             // 读锁
	rbuf       *buffer.Buffer         // 读缓存
	localAddr  string                 // 本地地址
	remoteAddr string                 // 远程地址
	status     int32                  // 当前状态
	client     bool                   // 用于标识是服务器端接收的连接还是客户端Dial产生的链接
	tag        string                 // 用于给外部标识类型
	data       map[string]interface{} // 自定义数据
	dataMux    sync.Mutex             // 保证数据安全
	readMux    sync.Mutex             // 读锁
}

// 基于标准的阻塞net.Conn实现,读写各起一个协程
type NetConn struct {
	Conn
	wbuf  *buffer.Buffer // 写缓存
	mutex sync.Mutex     // 锁
	cond  *sync.Cond     // 用于通知写协程退出
	sock  net.Conn       // 原始socket
}

func (c *Conn) Init(tran anet.Tran, client bool, tag string) {
	c.tran = tran
	c.client = client
	c.tag = tag
	c.rbuf = buffer.New()
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
	var result interface{}
	c.dataMux.Lock()
	if c.data != nil {
		result = c.data[key]
	}
	c.dataMux.Unlock()

	return result
}

func (c *Conn) Set(key string, val interface{}) {
	c.dataMux.Lock()
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	c.data[key] = val
	c.dataMux.Unlock()
}

func (c *Conn) SetStatus(s anet.Status) {
	atomic.StoreInt32(&c.status, int32(s))
}

func (c *Conn) Status() anet.Status {
	return anet.Status(atomic.LoadInt32(&c.status))
}

func (c *Conn) IsStatus(s anet.Status) bool {
	return c.Status() == s
}

func (c *Conn) IsActive() bool {
	return c.Status() == anet.OPEN
}

func (c *Conn) IsDial() bool {
	return c.client
}

func (c *Conn) LocalAddr() string {
	return c.localAddr
}

func (c *Conn) RemoteAddr() string {
	return c.remoteAddr
}

func (c *Conn) SetAddr(local string, remote string) {
	c.localAddr = local
	c.remoteAddr = remote
}

func (c *Conn) ReadLocker() sync.Locker {
	return &c.rmux
}

func (c *Conn) Read() *buffer.Buffer {
	return c.rbuf
}

/////////////////////////////////////////////////
// NetConn
/////////////////////////////////////////////////
func (c *NetConn) Init(tran anet.Tran, client bool, tag string) {
	c.Conn.Init(tran, client, tag)
	c.wbuf = buffer.New()
	c.cond = sync.NewCond(&c.mutex)
}

func (c *NetConn) Open(conn net.Conn) error {
	var err error
	c.mutex.Lock()
	if c.sock == nil {
		c.sock = conn
		c.SetAddr(conn.LocalAddr().String(), conn.RemoteAddr().String())
		c.SetStatus(anet.OPEN)
		c.rbuf.Clear()
	} else {
		err = anet.ErrHasOpened
	}

	c.mutex.Unlock()

	if err == nil {
		go c.doRead(conn)
		go c.doWrite()
		c.GetChain().HandleOpen(c)
	} else {
		c.GetChain().HandleError(c, err)
	}

	return err
}

func (c *NetConn) Close() error {
	c.mutex.Lock()
	status := c.Status()
	if status == anet.OPEN {
		c.SetStatus(anet.CLOSING)
	}
	c.mutex.Unlock()
	// 通知写线程退出
	// TODO:阻塞等待?
	if status == anet.OPEN {
		c.cond.Signal()
	}

	return nil
}

func (c *NetConn) Write(buf *buffer.Buffer) error {
	var err error
	c.mutex.Lock()
	// 连接过程中也可以发送,等连接成功后会主动发送所有数据
	// 如果连接失败则会清空数据
	status := c.Status()
	if status == anet.CONNECTING || status == anet.OPEN {
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

func (c *NetConn) Send(msg interface{}) error {
	c.tran.GetChain().HandleWrite(c, msg)
	return nil
}

func (c *NetConn) Error(err error) {
	// 清空连接前预发送的Buffer
	if c.IsStatus(anet.CONNECTING) {
		c.mutex.Lock()
		c.wbuf.Clear()
		c.mutex.Unlock()
	}
	c.GetChain().HandleError(c, err)
}

func (c *NetConn) doClose(err error) anet.Status {
	c.mutex.Lock()
	status := c.Status()
	if status == anet.CLOSED {
		c.mutex.Unlock()
		return status
	}

	c.wbuf.Clear()
	c.SetStatus(status)
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

func (c *NetConn) doRead(sock net.Conn) {
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

		c.rmux.Lock()
		c.rbuf.Append(data[:n])
		c.rmux.Unlock()

		c.GetChain().HandleRead(c, c.rbuf)
	}
}

func (c *NetConn) doWrite() {
	b := buffer.New()
	for {
		c.mutex.Lock()
		status := c.Status()
		for (status == anet.CONNECTING) || (status == anet.OPEN && c.wbuf.Empty()) {
			c.cond.Wait()
		}

		buffer.Swap(c.wbuf, b)
		sock := c.sock
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
