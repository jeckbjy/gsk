package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/jeckbjy/gsk/util/idgen/xid"

	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/util/addr"
)

func New(opts ...arpc.Option) arpc.Server {
	o := &arpc.Options{}
	for _, fn := range opts {
		fn(o)
	}

	return &_Server{opts: o}
}

type _Server struct {
	opts *arpc.Options
	addr string
}

func (s *_Server) Init(opts ...arpc.Option) error {
	for _, fn := range opts {
		fn(s.opts)
	}

	return nil
}

func (s *_Server) Start() error {
	o := s.opts
	if o.Id == "" {
		o.Id = xid.New().String()
	}

	// 监听服务
	l, err := o.Tran.Listen(o.Address)
	if err != nil {
		return err
	}
	s.addr = l.Addr().String()

	// 注册
	if err := s.register(); err != nil {
		return err
	}

	return nil
}

func (s *_Server) Stop() error {
	o := s.opts

	var gerr error
	if err := o.Tran.Close(); err != nil {
		gerr = err
	}

	if err := s.deregister(); err != nil {
		gerr = err
	}

	return gerr
}

func (s *_Server) register() error {
	o := s.opts
	if o.Registry == nil {
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
	srv := registry.NewService(o.Name, s.serviceID(), address, nil)
	// 解析endpoint
	if err := o.Registry.Register(srv); err != nil {
		return err
	}

	return nil
}

func (s *_Server) deregister() error {
	o := s.opts
	if o.Registry == nil {
		return nil
	}

	return o.Registry.Unregister(s.serviceID())
}

func (s *_Server) serviceID() string {
	o := s.opts
	return fmt.Sprintf("%s-%s", o.Name, o.Id)
}
