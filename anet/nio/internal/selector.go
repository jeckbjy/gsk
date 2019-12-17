package internal

import (
	"syscall"
)

func New() (*Selector, error) {
	poll := newPoller()
	if err := poll.Open(); err != nil {
		return nil, err
	}

	s := &Selector{poll: poll}
	return s, nil
}

type Selector struct {
	poll poller
}

func (s *Selector) Wakeup() error {
	return s.poll.Wakeup()
}

func (s *Selector) Wait(cb Callback) error {
	return s.poll.Wait(cb)
}

func (s *Selector) Add(fd uintptr) error {
	if err := syscall.SetNonblock(fd_t(fd), true); err != nil {
		return err
	}
	if err := s.poll.Add(fd); err != nil {
		return err
	}
	return nil
}

func (s *Selector) Delete(fd uintptr) error {
	return s.poll.Del(fd)
}

func (s *Selector) ModifyWrite(fd uintptr, add bool) error {
	return s.poll.ModifyWrite(fd, add)
}
