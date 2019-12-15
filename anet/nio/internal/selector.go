package internal

import "syscall"

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

func (s *Selector) Add(sock interface{}) error {
	fd, err := getFD(sock)
	if err != nil {
		return err
	}
	if err := syscall.SetNonblock(fd, true); err != nil {
		return err
	}
	if err := s.poll.Add(fd); err != nil {
		return err
	}
	return nil
}

func (s *Selector) Delete(sock interface{}) error {
	fd, err := getFD(sock)
	if err != nil {
		return err
	}
	return s.poll.Del(fd)
}

func (s *Selector) ModifyWrite(sock interface{}, add bool) error {
	fd, err := getFD(sock)
	if err != nil {
		return err
	}
	return s.poll.ModifyWrite(fd, add)
}
