// +build darwin netbsd freebsd openbsd dragonfly

package internal

import "syscall"

type FD = int

func SetReuseAddr(fd FD) {
	syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}
