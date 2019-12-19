// +build darwin dragonfly freebsd netbsd openbsd

package internal

import (
	"syscall"
)

type Kevent_t = syscall.Kevent_t

//const (
//	EVFILT_READ  = syscall.EVFILT_READ
//	EVFILT_WRITE = syscall.EVFILT_WRITE
//	EV_ADD       = syscall.EV_ADD
//	EV_DELETE    = syscall.EV_DELETE
//	EV_CLEAR     = syscall.EV_CLEAR
//)

func newPoller() Poller {
	return &kpoller{}
}

type kpoller struct {
	kfd    int                // kqueue fd
	events []syscall.Kevent_t // kqueue event
}

func (p *kpoller) IsSupportET() bool {
	return true
}

func (p *kpoller) Open() error {
	fd, err := syscall.Kqueue()
	if err != nil {
		return err
	}

	changes := []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}
	_, err = syscall.Kevent(fd, changes, nil, nil)
	if err != nil {
		_ = syscall.Close(fd)
		return err
	}

	syscall.CloseOnExec(fd)
	p.kfd = fd
	p.events = make([]syscall.Kevent_t, maxEventNum)
	return nil
}

func (p *kpoller) Close() error {
	return syscall.Close(p.kfd)
}

func (p *kpoller) Wakeup() error {
	changes := []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Fflags: syscall.NOTE_TRIGGER,
	}}
	_, err := syscall.Kevent(p.kfd, changes, nil, nil)
	return err
}

func (p *kpoller) Wait(cb Callback) error {
	pev := &Event{poll: p}
	for {
		n, err := syscall.Kevent(p.kfd, nil, p.events, nil)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
				continue
			}

			return err
		}

		for i := 0; i < n; i++ {
			kev := &p.events[i]
			pev.fd = FD(kev.Ident)
			pev.events = 0

			if kev.Flags&syscall.EV_ERROR != 0 {
				pev.events |= EventError
			} else if kev.Filter == syscall.EVFILT_READ {
				pev.events |= EventRead
			} else if kev.Filter == syscall.EVFILT_WRITE {
				pev.events |= EventWrite
			}

			cb(pev)
		}

		return nil
	}
}

// 添加并监听读事件,EV_CLEAR使用ET模式
func (p *kpoller) Add(fd FD) error {
	events := [1]Kevent_t{{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}
	_, err := syscall.Kevent(p.kfd, events[:], nil, nil)
	return err
}

// 删除读写监听事件
func (p *kpoller) Delete(fd FD) error {
	// 删除不存在的EVFILT_WRITE是否会有问题?
	events := [2]Kevent_t{
		{Ident: uint64(fd), Filter: syscall.EVFILT_READ, Flags: syscall.EV_DELETE},
		{Ident: uint64(fd), Filter: syscall.EVFILT_WRITE, Flags: syscall.EV_DELETE},
	}
	_, err := syscall.Kevent(p.kfd, events[:], nil, nil)
	return err
}

// 修改写监听
func (p *kpoller) ModifyWrite(fd FD, add bool) error {
	var flags uint16
	if add {
		flags = syscall.EV_ADD | syscall.EV_CLEAR
	} else {
		flags = syscall.EV_DELETE
	}

	events := [1]Kevent_t{{Ident: uint64(fd), Filter: syscall.EVFILT_READ, Flags: flags}}
	_, err := syscall.Kevent(p.kfd, events[:], nil, nil)
	return err
}
