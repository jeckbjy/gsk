package arpc

import (
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/addr"
)

func NewServer(opts ...ServerOption) IServer {
	s := &Server{opts: &ServerOptions{}}
	s.Init(opts...)
	return s
}

type Server struct {
	opts *ServerOptions
	addr string
}

func (s *Server) Options() ServerOptions {
	return *s.opts
}

func (s *Server) Init(opts ...ServerOption) {
	s.opts.Init(opts...)
}

func (s *Server) Start() error {
	s.opts.SetDefaults()
	conf := s.opts

	// before start hook
	if err := callHooks(s.opts.BeforeStart); err != nil {
		return err
	}

	// listen
	ts, err := conf.Tran.Listen(conf.Address)
	if err != nil {
		return err
	}

	s.addr = ts.Addr().String()

	// start broker
	// start register
	_ = s.Register()

	// after start hook
	if err := callHooks(s.opts.AfterStart); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	if err := callHooks(s.opts.BeforeStop); err != nil {
		return err
	}

	s.Deregister()

	if err := callHooks(s.opts.AfterStop); err != nil {
		return err
	}

	return nil
}

func (s *Server) Wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-s.opts.Context.Done():
	}
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	s.Wait()

	return s.Stop()
}

// Register 注册服务
func (s *Server) Register() error {
	conf := s.opts
	if !conf.RegistryEnable || conf.Registry == nil {
		return nil
	}

	var advt string
	if len(conf.Advertise) > 0 {
		advt = conf.Advertise
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

	opts := s.opts
	node := &registry.Node{
		Id:      opts.FullID(),
		Address: address,
	}

	srv := &registry.Service{
		Nodes: []*registry.Node{node},
	}

	rops := []registry.RegisterOption{
		registry.WithRegisterTTL(s.opts.RegisterTTL),
		registry.WithRegisterInterval(s.opts.RegisterInterval),
	}

	if err := opts.Registry.Register(srv, rops...); err != nil {
		return err
	}

	opts.Registry.Start()
	return nil
}

// Deregister 注销服务
func (s *Server) Deregister() {
	if s.opts.RegistryEnable || s.opts.Registry != nil {
		_ = s.opts.Registry.Deregister(s.opts.FullID())
		_ = s.opts.Registry.Stop()
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
