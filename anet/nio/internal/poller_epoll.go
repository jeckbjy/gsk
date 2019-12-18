// +build linux,!noepoll

package internal

import (
	"fmt"
	"syscall"
)

func newPoller() Poller {
	return &epoller{}
}

// https://medium.com/@copyconstruct/the-method-to-epolls-madness-d9d2d6378642
type epoller struct {
	efd    int // epoll fd
	wfd    int // wakeup fd
	events []syscall.EpollEvent
}

func (p *epoller) IsSupportET() bool {
	return true
}

func epoll_create() (int, error) {
	fd, err := syscall.EpollCreate1(0)
	if err == nil {
		return fd, nil
	}

	return syscall.EpollCreate(maxEventNum)
}

func (p *epoller) Open() error {
	fd, err := epoll_create()
	if err != nil {
		return err
	}

	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if e0 != 0 {
		_ = syscall.Close(fd)
		return fmt.Errorf("create eventfd fail")
	}

	syscall.CloseOnExec(fd)

	p.events = make([]syscall.EpollEvent, maxEventNum)
	p.efd = fd
	p.wfd = int(r0)
	return nil
}

func (p *epoller) Close() error {
	var err error
	if e := syscall.Close(p.wfd); e != nil {
		err = e
	}

	if e := syscall.Close(p.efd); e != nil {
		err = e
	}

	return err
}

func (p *epoller) Wakeup() error {
	_, err := syscall.Write(p.wfd, []byte{0, 0, 0, 0, 0, 0, 0, 1})
	return err
}

func (p *epoller) Wait(cb Callback) error {
	pev := &Event{poll: p}
	for {
		n, err := syscall.EpollWait(p.efd, p.events, 0)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
				continue
			}

			return err
		}

		for i := 0; i < n; i++ {
			ev := &p.events[i]
			fd := FD(ev.Fd)

			if fd == p.wfd {
				continue
			}

			pev.fd = fd
			pev.events = 0

			// https://stackoverflow.com/questions/24119072/how-to-deal-with-epollerr-and-epollhup/29206631
			// https://blog.csdn.net/halfclear/article/details/78061771?utm_source=blogxgwz8
			// from libev
			if ev.Events&(syscall.EPOLLIN|syscall.EPOLLERR|syscall.EPOLLHUP) != 0 {
				pev.events |= EventRead
			}

			if ev.Events&(syscall.EPOLLOUT|syscall.EPOLLERR|syscall.EPOLLHUP) != 0 {
				pev.events |= EventWrite
			}

			cb(pev)
		}

		return nil
	}
}

func (p *epoller) Add(fd FD) error {
	ev := &syscall.EpollEvent{Events: syscall.EPOLLIN | syscall.EPOLLET, Fd: int32(fd)}
	return syscall.EpollCtl(p.efd, syscall.EPOLL_CTL_ADD, fd, ev)
}

func (p *epoller) Delete(fd FD) error {
	return syscall.EpollCtl(p.efd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (p *epoller) ModifyWrite(fd FD, add bool) error {
	var events uint32
	if add {
		events = syscall.EPOLLIN | syscall.EPOLLOUT | syscall.EPOLLET
	} else {
		events = syscall.EPOLLIN | syscall.EPOLLET
	}

	ev := &syscall.EpollEvent{Events: events, Fd: int32(fd)}
	return syscall.EpollCtl(p.efd, syscall.EPOLL_CTL_MOD, fd, ev)
}
