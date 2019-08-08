package arpc

import (
	"fmt"
	"testing"
)

func TestDemo(t *testing.T) {
	type EchoReq struct {
		Text string `json:"text"`
	}
	type EchoRsp struct {
		Text string `json:"text"`
	}
	r := NewRouter()
	r.RegisterSrv(func(ctx IContext, req *EchoReq, rsp *EchoRsp) {
		t.Logf("clenet req %+v", req.Text)
		rsp.Text = "pong"
	})

	server := NewServer()
	go server.Run()

	count := 0
	client := NewClient()
	client.Call("echo", &EchoReq{Text: fmt.Sprintf("ping %+v", count)}, func(ctx IContext, rsp *EchoRsp) {
		t.Logf("server rsp %+v", rsp.Text)
		count++
		ctx.Send(&EchoReq{Text: fmt.Sprintf("ping %+v", count)})
	})
}
