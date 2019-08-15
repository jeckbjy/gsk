// +build darwin dragonfly freebsd netbsd openbsd

package internal

import "syscall"

func newPoller() poller {
	return &pkqueue{}
}

type pkqueue struct {
	kfd    int
	events []syscall.Kevent_t
}

func (p *pkqueue) Open() error {
	fd, err := syscall.Kqueue()
	if err != nil {
		return err
	}
	changes := []syscall.Kevent_t{{Ident: 0, Filter: syscall.EVFILT_USER, Flags: syscall.EV_ADD | syscall.EV_CLEAR}}
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

func (p *pkqueue) Close() error {
	return syscall.Close(p.kfd)
}

func (p *pkqueue) Wakeup() error {
	changes := []syscall.Kevent_t{{Ident: 0, Filter: syscall.EVFILT_USER, Fflags: syscall.NOTE_TRIGGER}}
	_, err := syscall.Kevent(p.kfd, changes, nil, nil)
	return err
}

func (p *pkqueue) Wait(s *Selector, cb SelectCB, msec int) error {
	for {
		n, err := syscall.Kevent(p.kfd, nil, p.events, nil)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
				continue
			}

			return err
		}
		for i := 0; i < n; i++ {
			ev := &p.events[i]
			fd := uintptr(ev.Ident)
			sk := s.keys[fd]
			if sk == nil {
				continue
			}

			if ev.Flags&(syscall.EV_ERROR|syscall.EV_EOF) != 0 {
				sk.setReadable()
				sk.setWritable()
				continue
			}

			if ev.Filter == syscall.EVFILT_READ {
				sk.setReadable()
			}
			if ev.Filter == syscall.EVFILT_WRITE {
				sk.setWritable()
			}

			if cb != nil {
				cb(sk)
			} else {
				s.readyKeys = append(s.readyKeys, sk)
			}
		}

		return nil
	}
}

func (p *pkqueue) Add(fd uintptr, ops Operation) error {
	changes := [4]syscall.Kevent_t{}
	num := p.control(&changes, 0, fd, ops, true)
	_, err := syscall.Kevent(p.kfd, changes[:num], nil, nil)
	return err
}

func (p *pkqueue) Del(fd uintptr, ops Operation) error {
	changes := [4]syscall.Kevent_t{}
	num := p.control(&changes, 0, fd, ops, false)
	_, err := syscall.Kevent(p.kfd, changes[:num], nil, nil)
	return err
}

func (p *pkqueue) Mod(fd uintptr, old, ops Operation) error {
	changes := [4]syscall.Kevent_t{}
	num := 0

	if old != 0 {
		// delete old
		num = p.control(&changes, 0, fd, old, false)
	}

	num = p.control(&changes, num, fd, ops, true)
	_, err := syscall.Kevent(p.kfd, changes[:num], nil, nil)
	return err
}

func (p *pkqueue) control(changes *[4]syscall.Kevent_t, num int, fd uintptr, ops Operation, add bool) int {
	ident := uint64(fd)
	var flags uint16
	if add {
		flags = syscall.EV_ADD | syscall.EV_CLEAR
	} else {
		flags = syscall.EV_DELETE
	}

	if ops&OP_READ != 0 {
		changes[num] = syscall.Kevent_t{Filter: syscall.EVFILT_READ, Ident: ident, Flags: flags}
		num++
	}

	if ops&OP_WRITE != 0 {
		changes[num] = syscall.Kevent_t{Filter: syscall.EVFILT_WRITE, Ident: ident, Flags: flags}
		num++
	}

	return num
}
