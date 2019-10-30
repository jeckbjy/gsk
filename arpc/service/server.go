package service

import (
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/addr"

	"github.com/jeckbjy/gsk/arpc"
)

type Server struct {
	opts *arpc.Options
	addr string
}

func (s *Server) Register(srv interface{}) {

}

func (s *Server) Start() error {
	o := s.opts

	// before start
	if err := callHooks(s.opts.BeforeStart); err != nil {
		return err
	}

	// listen
	l, err := o.Tran.Listen(o.Address)
	if err != nil {
		return err
	}
	s.addr = l.Addr().String()

	// register service
	if err := s.register(); err != nil {
		return err
	}

	// after start
	if err := callHooks(s.opts.AfterStart); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	o := s.opts

	var gerr error
	if err := callHooks(o.BeforeStop); err != nil {
		gerr = err
	}

	// close listener
	if err := o.Tran.Close(); err != nil {
		gerr = err
	}

	// deregister

	if err := callHooks(o.AfterStop); err != nil {
		gerr = err
	}

	return gerr
}

func (s *Server) Wait() {
	o := s.opts

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	if o.Context != nil {
		select {
		case <-ch:
		case <-o.Context.Done():
		}
	} else {
		select {
		case <-ch:
		}
	}
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	s.Wait()
	return s.Stop()
}

func (s *Server) register() error {
	o := s.opts
	if o.HasFlag(arpc.DisableRegistry) {
		return nil
	}

	var advt string
	if len(o.Advertise) > 0 {
		advt = o.Advertise
	} else {
		advt = s.addr
	}

	var host, port string
	if strings.LastIndexByte(advt, ':') != -1 {
		var err error
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = advt
	}

	address, err := addr.Extract(host)
	if err != nil {
		return err
	}

	address = net.JoinHostPort(host, port)

	node := &registry.Node{
		Id:      o.FullId(),
		Address: address,
	}

	srv := &registry.Service{
		Nodes: []*registry.Node{node},
	}

	if err := o.Registry.Register(srv); err != nil {
		return err
	}

	o.Registry.Start()
	return nil
}

func (s *Server) deregister() {
	o := s.opts
	if !o.HasFlag(arpc.DisableRegistry) {
		_ = o.Registry.Deregister(o.FullId())
		_ = o.Registry.Stop()
	}
}

func callHooks(hooks []func() error) error {
	for _, fn := range hooks {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
