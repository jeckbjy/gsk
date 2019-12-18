// +build windows

package internal

import "syscall"

type FD = syscall.Handle

func SetReuseAddr(fd FD) {
	syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}
