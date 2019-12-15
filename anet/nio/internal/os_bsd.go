// +build darwin netbsd freebsd openbsd dragonfly

package internal

import "syscall"

type fd_t = int

const (
	evRead  = syscall.EVFILT_READ
	evWrite = syscall.EVFILT_WRITE
	evError = syscall.EV_ERROR
)
