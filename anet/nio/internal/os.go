package internal

import (
	"fmt"
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
func GetFD(conn interface{}) (uintptr, error) {
	if i, ok := conn.(iFile); ok {
		if f, err := i.File(); err == nil {
			return f.Fd(), err
		} else {
			return 0, err
		}
	}

	return 0, fmt.Errorf("bad file descriptor")
}

func SetNonblock(fd uintptr) error {
	return syscall.SetNonblock(fd_t(fd), true)
}

func Read(fd uintptr, p []byte) (int, error) {
	return syscall.Read(fd_t(fd), p)
}

func Write(fd uintptr, p []byte) (int, error) {
	return syscall.Write(fd_t(fd), p)
}
