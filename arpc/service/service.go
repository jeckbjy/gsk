package service

import (
	"github.com/jeckbjy/gsk/arpc"
)

func New(opts *arpc.Options) (arpc.Service, error) {
	s := &Service{}
	if err := s.Init(opts); err != nil {
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

func (s *Service) Init(opts *arpc.Options) error {
	s.opts = opts
	s.Server.opts = opts
	s.Client.opts = opts
	return nil
}
