package internal

import (
	"fmt"
	"os"
	"syscall"
)

// iFile describes an object that has ability to return os.File.
type iFile interface {
	// File returns a copy of object's file descriptor.
	File() (*os.File, error)
}

func GetFd(conn interface{}) (uintptr, error) {
	if i, ok := conn.(iFile); ok {
		if f, err := i.File(); err == nil {
			return f.Fd(), err
		} else {
			return 0, err
		}
	}

	return 0, fmt.Errorf("bad file descriptor")
}

func SetNonblock(fd uintptr, nonblocking bool) error {
	return syscall.SetNonblock(Handle(fd), nonblocking)
}

func Read(fd uintptr, p []byte) (n int, err error) {
	return syscall.Read(Handle(fd), p)
}

func Write(fd uintptr, p []byte) (n int, err error) {
	return syscall.Write(Handle(fd), p)
}
