// +build windows

package internal

import (
	"syscall"
	"time"
)

type FD = syscall.Handle

func SetReuseAddr(fd FD) {
	syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}

//sys _select(nfds int, readfds *FDSet, writefds *FDSet, exceptfds *FDSet, timeout *syscall.Timeval) (total int, err error) = ws2_32.select
//sys __WSAFDIsSet(handle syscall.Handle, fdset *FDSet) (isset int, err error) = ws2_32.__WSAFDIsSet

func sysSelect(n int, r, w, e *fdSet, timeout *syscall.Timeval) (int, error) {
	return _select(n, r, w, e, timeout)
}

func goSelect(n int, r, w, e *fdSet, timeout time.Duration) (int, error) {
	var timeval *syscall.Timeval
	if timeout >= 0 {
		t := syscall.NsecToTimeval(timeout.Nanoseconds())
		timeval = &t
	}

	return sysSelect(n, r, w, e, timeval)
}

const fdset_SIZE = 64

type fdSet struct {
	fd_count uint
	fd_array [fdset_SIZE]uintptr
}

// Set adds the fd to the set
func (fds *fdSet) Set(fd uintptr) {
	var i uint
	for i = 0; i < fds.fd_count; i++ {
		if fds.fd_array[i] == fd {
			break
		}
	}
	if i == fds.fd_count {
		if fds.fd_count < fdset_SIZE {
			fds.fd_array[i] = fd
			fds.fd_count++
		}
	}
}

// Clear remove the fd from the set
func (fds *fdSet) Clear(fd uintptr) {
	var i uint
	for i = 0; i < fds.fd_count; i++ {
		if fds.fd_array[i] == fd {
			for i < fds.fd_count-1 {
				fds.fd_array[i] = fds.fd_array[i+1]
				i++
			}
			fds.fd_count--
			break
		}
	}
}

// IsSet check if the given fd is set
func (fds *fdSet) IsSet(fd uintptr) bool {
	if isset, err := __WSAFDIsSet(syscall.Handle(fd), fds); err == nil && isset != 0 {
		return true
	}
	return false
}

// Zero empties the Set
func (fds *fdSet) Zero() {
	fds.fd_count = 0
}
