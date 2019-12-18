package nio

import (
	"log"
	"syscall"

	"github.com/jeckbjy/gsk/anet/nio/internal"
)

func newListener(listener *internal.Listener, poller internal.Poller, tran *nioTran, tag string) (*nioListener, error) {
	l := &nioListener{Listener: listener, poller: poller, tran: tran, tag: tag}
	if err := l.Open(); err != nil {
		return nil, err
	}
	return l, nil
}

type nioListener struct {
	*internal.Listener
	poller internal.Poller // nio selector
	tran   *nioTran
	tag    string
}

func (l *nioListener) onEvent(*internal.Event) {
	for {
		sock, err := l.Accept()
		if err != nil {
			if err != syscall.EAGAIN {
				log.Printf("accept fail,%+v", err)
			}
			break
		}
		poller := gLoop.next()
		if poller == nil {
			_ = sock.Close()
			break
		}
		conn := newConn(l.tran, false, l.tag, poller)
		_ = conn.Open(sock)
	}
}

func (l *nioListener) Open() error {
	gLoop.add(l.Fd(), l)
	if err := l.poller.Add(l.Fd()); err != nil {
		_ = l.Close()
		return err
	}

	return nil
}

func (l *nioListener) Close() error {
	_ = l.poller.Delete(l.Fd())
	gLoop.remove(l.Fd())
	return l.Listener.Close()
}
