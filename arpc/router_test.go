package arpc

import (
	"fmt"
	"testing"

	"github.com/jeckbjy/gsk/arpc/codec/jsonc"
	"github.com/jeckbjy/gsk/util/buffer"
)

type EchoReq struct {
}

type EchoRsp struct {
}

func HandleEcho(ctx IContext, req *EchoReq) error {
	fmt.Printf("HandleEcho:req")
	return nil
}

type EchoService struct {
}

func (s *EchoService) Handle(ctx IContext, req *EchoReq, rsp *EchoRsp) error {
	fmt.Printf("EchoService handle:req")
	return nil
}

func Find(t *testing.T, r IRouter, pkg IPacket) {
	h, err := r.Find(pkg)
	if err != nil {
		t.Error(err)
	} else if h != nil {
		pkg.SetCodec(jsonc.New())
		pkg.SetBytes(buffer.New())
		ctx := NewContext(nil, pkg, NewPacket())
		_ = h(ctx)
	} else {
		t.Error("not found handler")
	}
}

func FindMethod(t *testing.T, r IRouter, method string) {
	p := NewPacket()
	p.SetMethod(method)
	Find(t, r, p)
}

func FindName(t *testing.T, r IRouter, name string) {
	p := NewPacket()
	p.SetName(name)
	Find(t, r, p)
}

func TestRouter(t *testing.T) {
	//1: register function
	r1 := NewRouter()
	if err := r1.RegisterSrv(HandleEcho); err != nil {
		t.Error(err)
	} else {
		FindMethod(t, r1, "HandleEcho")
		FindName(t, r1, "EchoReq")
	}

	// 2: register service
	r2 := NewRouter()
	if err := r2.RegisterSrv(&EchoService{}); err != nil {
		t.Error(err)
	} else {
		FindMethod(t, r2, "Handle")
		FindName(t, r2, "EchoReq")
	}

	// 3: register message
	//if err := r.Register(&EchoRsp{}); err != nil {
	//	t.Error(err)
	//}

	// 4: register rpc?

	// test find?
}
