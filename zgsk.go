package gsk

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/tcp"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/frame"
	"github.com/jeckbjy/gsk/arpc/frame/varint"
	"github.com/jeckbjy/gsk/arpc/router"
	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/exec/simple"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/registry/local"
)

// 设置默认参数
func init() {
	anet.SetDefault(tcp.New)
	registry.SetDefault(local.New(nil))
	frame.SetDefault(varint.New())
	exec.SetDefault(simple.New())
	arpc.SetDefaultRouter(router.New())
}

// 强制要求提供服务名
func New(name string, opts ...Option) Service {
	o := newOptions(name, opts...)
	srv := &service{}
	srv.Init(o)
	return srv
}

// Service 聚合各个微服务组件,方便外部调用
type Service interface {
	Name() string
	Run() error
	Register(callback interface{}, opts ...arpc.RegisterOption) error
	Send(service string, req interface{}, opts ...arpc.CallOption) error
	Call(service string, req interface{}, rsp interface{}, opts ...arpc.CallOption) error
}
