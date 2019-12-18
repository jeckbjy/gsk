package internal

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

const EAGAIN = syscall.EAGAIN

// iFile describes an object that has ability to return os.File.
type iFile interface {
	// File returns a copy of object's file descriptor.
	File() (*os.File, error)
}

// 同一个conn,每次调用都会自增1,还不能随意调用
func getFD(conn interface{}) (uintptr, error) {
	if i, ok := conn.(iFile); ok {
		if f, err := i.File(); err == nil {
			return f.Fd(), err
		} else {
			return 0, err
		}
	}

	return 0, fmt.Errorf("bad file descriptor")
}

func SetNonblock(fd FD) error {
	return syscall.SetNonblock(fd, true)
}

func getSockaddr(network, address string) (syscall.Sockaddr, int, error) {
	addr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, -1, err
	}

	switch network {
	case "tcp", "tcp4":
		var sa4 syscall.SockaddrInet4
		sa4.Port = addr.Port
		copy(sa4.Addr[:], addr.IP.To4())
		return &sa4, syscall.AF_INET, nil
	case "tcp6":
		var sa6 syscall.SockaddrInet6
		sa6.Port = addr.Port
		copy(sa6.Addr[:], addr.IP.To16())
		if addr.Zone != "" {
			ifi, err := net.InterfaceByName(addr.Zone)
			if err != nil {
				return nil, -1, err
			}
			sa6.ZoneId = uint32(ifi.Index)
		}
		return &sa6, syscall.AF_INET6, nil
	default:
		return nil, -1, errors.New("Unknown network type " + network)
	}
}

// getNetAddr returns a go/net friendly address
func getNetAddr(sa syscall.Sockaddr) net.Addr {
	var a net.Addr
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
		}
	case *syscall.SockaddrInet6:
		var zone string
		if sa.ZoneId != 0 {
			if ifi, err := net.InterfaceByIndex(int(sa.ZoneId)); err == nil {
				zone = ifi.Name
			}
		}
		if zone == "" && sa.ZoneId != 0 {
		}
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
			Zone: zone,
		}
	case *syscall.SockaddrUnix:
		a = &net.UnixAddr{Net: "unix", Name: sa.Name}
	}
	return a
}

func Listen(network, address string) (*Listener, error) {
	sa, st, err := getSockaddr(network, address)
	if err != nil {
		return nil, err
	}

	// create socket
	fd, err := syscall.Socket(st, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	// 不设置close后会产生一段时间的TIME_WAIT
	SetReuseAddr(fd)

	// bind
	if err := syscall.Bind(fd, sa); err != nil {
		return nil, err
	}

	if err := syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		return nil, err
	}

	return newListener(fd, sa), nil
}

// 不支持timeout
func Dial(network, address string) (*Conn, error) {
	sa, st, err := getSockaddr(network, address)
	if err != nil {
		return nil, err
	}

	fd, err := syscall.Socket(st, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	if err := syscall.Connect(fd, sa); err != nil {
		return nil, err
	}

	if err := SetNonblock(fd); err != nil {
		return nil, syscall.Close(fd)
	}

	return newConn(fd, sa), nil
}
