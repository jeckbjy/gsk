// +build linux

package internal

import (
	"syscall"
)

type fd_t = int

const (
	evRead  = syscall.EPOLLIN
	evWrite = syscall.EPOLLOUT
	evError = syscall.EPOLLERR
)
