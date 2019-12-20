package gsk

import (
	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/tcp"
	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/arpc/router"
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/codec/gobc"
	"github.com/jeckbjy/gsk/codec/jsonc"
	"github.com/jeckbjy/gsk/codec/protoc"
	"github.com/jeckbjy/gsk/codec/xmlc"
	"github.com/jeckbjy/gsk/exec"
	"github.com/jeckbjy/gsk/exec/simple"
	"github.com/jeckbjy/gsk/frame"
	"github.com/jeckbjy/gsk/frame/varint"
	"github.com/jeckbjy/gsk/registry"
	"github.com/jeckbjy/gsk/registry/local"
)

// 设置默认参数
func init() {
	codec.Add(xmlc.New())
	codec.Add(gobc.New())
	codec.Add(jsonc.New())
	codec.Add(protoc.New())

	anet.SetDefault(tcp.New)
	registry.SetDefault(local.New())
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
