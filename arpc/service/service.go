package service

import (
	"github.com/jeckbjy/gsk/arpc"
)

func New(opts ...arpc.Option) (arpc.Service, error) {
	o := &arpc.Options{}
	s := &Service{opts: o}
	s.Server.opts = o
	s.Client.opts = o

	if err := s.Init(opts...); err != nil {
		return nil, err
	}

	return s, nil
}

type Service struct {
	opts *arpc.Options
	Server
	Client
}

func (s *Service) Options() *arpc.Options {
	return s.opts
}

func (s *Service) Init(opts ...arpc.Option) error {
	for _, fn := range opts {
		fn(s.opts)
	}

	return nil
}
