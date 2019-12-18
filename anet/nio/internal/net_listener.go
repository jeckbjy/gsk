package internal

import (
	"net"
	"syscall"
)

func newListener(fd FD, addr syscall.Sockaddr) *Listener {
	return &Listener{fd: fd, addr: getNetAddr(addr)}
}

type Listener struct {
	fd   FD
	addr net.Addr
}

func (l *Listener) Fd() FD {
	return l.fd
}

func (l *Listener) Addr() net.Addr {
	return l.addr
}

func (l *Listener) Accept() (*Conn, error) {
	nfd, sa, err := syscall.Accept(l.fd)
	if err != nil {
		return nil, err
	}

	if err := SetNonblock(nfd); err != nil {
		_ = syscall.Close(nfd)
		return nil, err
	}

	return newConn(nfd, sa), nil
}

func (l *Listener) Close() error {
	return syscall.Close(l.fd)
}
