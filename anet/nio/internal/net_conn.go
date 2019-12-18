package internal

import (
	"net"
	"syscall"
	"time"
)

func newConn(fd FD, remote syscall.Sockaddr) *Conn {
	sa, err := syscall.Getsockname(fd)
	var local net.Addr
	if err == nil {
		local = getNetAddr(sa)
	} else {
		local = &net.TCPAddr{}
	}
	return &Conn{fd: fd, local: local, remote: getNetAddr(remote)}
}

type Conn struct {
	fd     FD
	local  net.Addr
	remote net.Addr
}

func (c *Conn) Fd() FD {
	return c.fd
}

func (c *Conn) Read(p []byte) (int, error) {
	return syscall.Read(c.fd, p)
}

func (c *Conn) Write(p []byte) (int, error) {
	return syscall.Write(c.fd, p)
}

func (c *Conn) Close() error {
	return syscall.Close(c.fd)
}

func (c *Conn) LocalAddr() net.Addr {
	return c.local
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *Conn) SetDeadline(t time.Time) error {
	return ErrNotSupport
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return ErrNotSupport
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return ErrNotSupport
}
