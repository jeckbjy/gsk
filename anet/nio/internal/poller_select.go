// +build windows

package internal

import (
	"os"
	"syscall"
)

func newPoller() Poller {
	return &spoller{}
}

type spoller struct {
	fdmax FD
	fds   []FD
	rset  fdSet
	wset  fdSet
	eset  fdSet
	pr    *os.File
	pw    *os.File
}

func (p *spoller) IsSupportET() bool {
	return false
}

func (p *spoller) Open() error {
	p.fdmax = -1
	p.rset.Zero()
	p.wset.Zero()
	p.eset.Zero()
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	p.pr = r
	p.pw = w
	p.rset.Set(r.Fd())
	p.fdmax = FD(r.Fd())
	p.fds = append(p.fds, FD(r.Fd()))

	return nil
}

func (p *spoller) Close() error {
	e1 := p.pr.Close()
	e2 := p.pw.Close()
	if e1 != nil {
		return e1
	}
	return e2
}

func (p *spoller) Wakeup() error {
	_, err := p.pw.Write([]byte("0"))
	return err
}

func (p *spoller) Wait(cb Callback) error {
	pev := &Event{poll: p}
	p.eset.Zero()
	for {
		_, err := goSelect(int(p.fdmax+1), &p.rset, &p.wset, &p.eset, -1)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
				continue
			}

			return err
		}

		// test wakeup
		if p.rset.IsSet(p.pr.Fd()) {
			// drain all
			bytes := [64]byte{}
			for {
				_, err := p.pr.Read(bytes[:64])
				if err != nil {
					if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
						continue
					}
					break
				}
			}
		}

		// for loop
		for _, fd := range p.fds {
			if p.eset.IsSet(uintptr(fd)) {
				pev.events |= EventError
				cb(pev)
				continue
			}

			if p.rset.IsSet(uintptr(fd)) {
				pev.events |= EventRead
			}
			if p.wset.IsSet(uintptr(fd)) {
				pev.events |= EventWrite
			}

			if pev.events != 0 {
				cb(pev)
			}
		}

		return nil
	}
}

func (p *spoller) Add(fd FD) error {
	if fd >= p.fdmax {
		p.fdmax = fd
	}
	p.rset.Set(uintptr(fd))
	p.fds = append(p.fds, fd)
	return nil
}

func (p *spoller) Delete(fd FD) error {
	p.rset.Clear(uintptr(fd))
	p.wset.Clear(uintptr(fd))
	// get max fd
	fdmax := FD(0)
	index := -1
	for i, f := range p.fds {
		if f == fd {
			index = i
		} else if f > fdmax {
			fdmax = f
		}
	}

	p.fdmax = fdmax
	if index != -1 {
		p.fds = append(p.fds[:index], p.fds[index+1:]...)
	}

	return nil
}

func (p *spoller) ModifyWrite(fd FD, add bool) error {
	if add {
		p.wset.Set(uintptr(fd))
	} else {
		p.wset.Clear(uintptr(fd))
	}

	return nil
}
