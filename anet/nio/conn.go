package nio

import (
	"errors"
	"log"
	"sync"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/base"
	"github.com/jeckbjy/gsk/anet/nio/internal"
	"github.com/jeckbjy/gsk/util/buffer"
)

var errPeerClosed = errors.New("socket closed")

func newConn(tran anet.Tran, client bool, tag string, poller internal.Poller) *nioConn {
	conn := &nioConn{}
	conn.Init(tran, client, tag, poller)
	return conn
}

type nioConn struct {
	base.Conn
	mux     sync.Mutex
	sock    *internal.Conn
	poller  internal.Poller // nio selector
	wbuf    *buffer.Buffer  // 写缓存
	writing bool            // 标识当前是否监听写事件
}

func (c *nioConn) Init(tran anet.Tran, client bool, tag string, poller internal.Poller) {
	c.Conn.Init(tran, client, tag)
	c.wbuf = buffer.New()
	c.poller = poller
}

func (c *nioConn) Open(sock *internal.Conn) error {
	var err error
	c.mux.Lock()
	if c.sock == nil {
		err = c.poller.Add(sock.Fd())
		if err == nil {
			gLoop.add(sock.Fd(), c)
			c.sock = sock
			c.writing = false
			c.SetAddr(sock.LocalAddr().String(), sock.RemoteAddr().String())
			c.SetStatus(anet.OPEN)
			if !c.wbuf.Empty() {
				_ = c.doWrite()
			}
		}
	} else {
		err = anet.ErrHasOpened
	}
	c.mux.Unlock()

	if err == nil {
		c.GetChain().HandleOpen(c)
	} else {
		c.onError(err)
	}

	return err
}

func (c *nioConn) Close() error {
	c.mux.Lock()
	if c.Status() == anet.OPEN {
		if c.wbuf.Empty() {
			// 直接关闭并清理所有数据
			c.doClose()
		} else {
			// 等待所有数据发送完再退出?
			c.SetStatus(anet.CLOSING)
		}
	}
	c.mux.Unlock()
	return nil
}

func (c *nioConn) Send(msg interface{}) error {
	c.Tran().GetChain().HandleWrite(c, msg)
	return nil
}

// 写数据,直到写完成
func (c *nioConn) Write(data *buffer.Buffer) error {
	c.mux.Lock()
	var err error
	status := c.Status()

	switch status {
	case anet.CONNECTING:
		c.wbuf.AppendBuffer(data)
	case anet.OPEN:
		if c.wbuf.Empty() {
			buffer.Swap(c.wbuf, data)
			err = c.doWrite()
			if err != nil {
				c.doClose()
			}
		} else {
			c.wbuf.AppendBuffer(data)
		}
	default:
		err = anet.ErrHasClosed
	}

	c.mux.Unlock()
	return err
}

func (c *nioConn) modifyWrite(writing bool) {
	if c.writing != writing {
		c.writing = writing
		_ = c.poller.ModifyWrite(c.sock.Fd(), writing)
	}
}

// 发送则要全部发送完,直到不能发送为止
func (c *nioConn) doWrite() error {
	log.Printf("doWrite")
	iter := c.wbuf.Iter()
	for iter.Next() {
		data := iter.Data()
		n, err := c.sock.Write(data)
		if n < len(data) {
			if n == -1 && err != internal.EAGAIN {
				// 发生错误
				return err
			}

			// 删除已经发送的数据
			if n > 0 {
				_, _ = c.wbuf.Seek(int64(n), buffer.SeekStart)
				c.wbuf.Discard()
			}
			//
			c.modifyWrite(true)
			break
		}

		// 删除已经发送的数据
		iter.Remove()
	}

	return nil
}

// 读取则要全部读完,直到不能读取为止
func (c *nioConn) doRead() error {
	var result error
	reader := c.Read()
	rmux := c.ReadLocker()
	rmux.Lock()
	for {
		data := make([]byte, 1024)
		n, err := c.sock.Read(data)
		if n < 0 {
			if err != internal.EAGAIN {
				log.Printf("read, %+v,%+v", n, err)
				result = err
			}
			break
		}

		if n == 0 {
			// 对方关闭了连接?epoll可以这样检测,kqueue可以么？
			err = errPeerClosed
			break
		}

		if n == len(data) {
			reader.Append(data)
		} else {
			reader.Append(data[:n])
		}
	}
	rmux.Unlock()

	return result
}

func (c *nioConn) doClose() {
	status := c.Status()
	if status == anet.CLOSED {
		return
	}

	c.wbuf.Clear()
	if c.sock != nil {
		gLoop.remove(c.sock.Fd())
		_ = c.poller.Delete(c.sock.Fd())
		_ = c.sock.Close()
		c.sock = nil
	}
}

func (c *nioConn) onEvent(ev *internal.Event) {
	if ev.HasError() {
		c.mux.Lock()
		c.doClose()
		c.mux.Unlock()
		return
	}

	if ev.Readable() {
		if err := c.doRead(); err == nil {
			c.GetChain().HandleRead(c, c.Read())
		} else {
			c.mux.Lock()
			c.doClose()
			c.mux.Unlock()
			if err != errPeerClosed {
				c.onError(err)
			}
			return
		}
	}

	if ev.Writable() {
		var err error
		c.mux.Lock()
		if c.wbuf.Empty() {
			// 没有需要发送的内容,但是收到了发送事件,bug?
			c.modifyWrite(false)
		} else {
			err = c.doWrite()
		}

		if err != nil || (c.Status() == anet.CLOSING && c.wbuf.Empty()) {
			c.doClose()
		}
		c.mux.Unlock()

		if err != nil {
			c.onError(err)
		}
	}
}

func (c *nioConn) onError(err error) {
	c.GetChain().HandleError(c, err)
}
