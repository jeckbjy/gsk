package gsk

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeckbjy/gsk/apm/alog"
	"github.com/jeckbjy/gsk/arpc"
)

type service struct {
	context     context.Context
	server      arpc.Server
	client      arpc.Client
	beforeStart []Callback
	afterStart  []Callback
	beforeStop  []Callback
	afterStop   []Callback
	router      arpc.Router
	name        string
	exitCh      chan os.Signal
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Init(opts *Options) {
	s.server = opts.Server
	s.client = opts.Client
	s.context = opts.Context
	s.beforeStart = opts.BeforeStart
	s.afterStart = opts.AfterStart
	s.beforeStop = opts.BeforeStop
	s.afterStop = opts.AfterStop
	s.router = opts.Router
	s.name = opts.Name
}

func (s *service) Start() error {
	for _, fn := range s.beforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	if err := s.server.Start(); err != nil {
		return err
	}

	for _, fn := range s.afterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) Stop() error {
	var result error
	for _, fn := range s.beforeStop {
		if err := fn(); err != nil {
			result = err
		}
	}

	if err := s.server.Stop(); err != nil {
		result = err
	}

	for _, fn := range s.afterStop {
		if err := fn(); err != nil {
			result = err
		}
	}

	return result
}

func (s *service) Run() error {
	if err := s.Start(); err != nil {
		alog.Error(err)
		return err
	}

	log.Printf("service run")

	// wait
	ch := make(chan os.Signal, 1)
	s.exitCh = ch
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-s.context.Done():
	}

	if err := s.Stop(); err != nil {
		alog.Error(err)
		return err
	}

	return nil
}

func (s *service) Exit() {
	if s.exitCh != nil {
		s.exitCh <- syscall.SIGQUIT
	}
}

func (s *service) Register(callback interface{}, opts ...arpc.RegisterOption) error {
	return s.router.Register(callback, opts...)
}

func (s *service) Send(service string, req interface{}, opts ...arpc.CallOption) error {
	return s.client.Send(service, req, opts...)
}

func (s *service) Call(service string, req interface{}, rsp interface{}, opts ...arpc.CallOption) error {
	return s.client.Call(service, req, rsp, opts...)
}
