// +build linux

package internal

import (
	"syscall"
)

type FD = int

const so_REUSEPORT = 0xf

func SetReuseAddr(fd FD) {
	syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, so_REUSEPORT, 1)
}
