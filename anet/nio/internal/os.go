package internal

import (
	"fmt"
	"os"
)

// iFile describes an object that has ability to return os.File.
type iFile interface {
	// File returns a copy of object's file descriptor.
	File() (*os.File, error)
}

func getFD(conn interface{}) (fd_t, error) {
	if i, ok := conn.(iFile); ok {
		if f, err := i.File(); err == nil {
			return fd_t(f.Fd()), err
		} else {
			return 0, err
		}
	}

	return 0, fmt.Errorf("bad file descriptor")
}
