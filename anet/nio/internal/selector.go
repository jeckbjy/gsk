package internal

import "fmt"

var (
	// ErrRegistered is returned by Poller Start() method to indicate that
	// connection with the same underlying file descriptor was already
	// registered within the poller instance.
	ErrRegistered = fmt.Errorf("file descriptor is already registered in poller instance")

	// ErrNotRegistered is returned by Poller Stop() and Resume() methods to
	// indicate that connection with the same underlying file descriptor was
	// not registered before within the poller instance.
	ErrNotRegistered = fmt.Errorf("file descriptor was not registered before in poller instance")

	// ErrNotSupport not support
	ErrNotSupport = fmt.Errorf("not support")
)

func New() (*Selector, error) {
	poll := newPoller()
	if poll == nil {
		return nil, fmt.Errorf("create poller fail")
	}

	if err := poll.Open(); err != nil {
		return nil, err
	}

	s := &Selector{keys: make(map[uintptr]*SelectionKey), poll: poll}
	return s, nil
}

type Selector struct {
	poll      poller
	keys      map[uintptr]*SelectionKey
	readyKeys []*SelectionKey
}

func (s *Selector) Select(ops ...SelectOption) ([]*SelectionKey, error) {
	conf := SelectOptions{}
	conf.Timeout = -1
	for _, fn := range ops {
		fn(&conf)
	}
	s.readyKeys = s.readyKeys[:0]
	if err := s.poll.Wait(s, conf.Callback, conf.Timeout); err != nil {
		return nil, err
	}

	return s.readyKeys, nil
}

func (s *Selector) Wakeup() error {
	return s.poll.Wakeup()
}

func (s *Selector) Add(sock interface{}, ops Operation, data interface{}) (*SelectionKey, error) {
	fd, err := GetFd(sock)
	if err != nil {
		return nil, err
	}

	if s.keys[fd] != nil {
		return nil, ErrRegistered
	}

	if err := SetNonblock(fd, true); err != nil {
		return nil, err
	}

	if err := s.poll.Add(fd, ops); err != nil {
		return nil, err
	}

	sk := &SelectionKey{sock: sock, data: data, fd: fd, interests: ops}
	s.keys[fd] = sk
	return sk, nil
}

func (s *Selector) Delete(channel interface{}) error {
	sk, err := s.getSelectionKey(channel)
	if err != nil {
		return err
	}

	delete(s.keys, sk.fd)
	return s.poll.Del(sk.fd, sk.interests)
}

func (s *Selector) Modify(channel interface{}, ops Operation) error {
	sk, err := s.getSelectionKey(channel)
	if err != nil {
		return err
	}

	if ops == sk.interests {
		return nil
	}

	old := sk.interests
	sk.interests = ops

	return s.poll.Mod(sk.fd, old, ops)
}

// ModifyXOR 切换某个状态开关,通常用于OP_WRITE状态控制
func (s *Selector) ModifyXOR(channel interface{}, ops Operation) error {
	if ops == 0 {
		return nil
	}

	sk, err := s.getSelectionKey(channel)
	if err != nil {
		return err
	}

	old := sk.interests
	sk.interests ^= ops

	return s.poll.Mod(sk.fd, old, sk.interests)
}

// ModifyIf 添加或删除某个状态
func (s *Selector) ModifyIf(channel interface{}, ops Operation, add bool) error {
	if ops == 0 {
		return nil
	}

	sk, err := s.getSelectionKey(channel)
	if err != nil {
		return err
	}

	old := sk.interests
	if add {
		sk.interests |= ops
	} else {
		sk.interests &^= ops
	}

	return s.poll.Mod(sk.fd, old, sk.interests)
}

func (s *Selector) getSelectionKey(channel interface{}) (*SelectionKey, error) {
	fd, err := GetFd(channel)
	if err != nil {
		return nil, err
	}

	sk := s.keys[fd]
	if sk == nil {
		return nil, ErrNotRegistered
	}

	return sk, nil
}
