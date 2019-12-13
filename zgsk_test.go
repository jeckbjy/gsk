package gsk

import (
	"fmt"
	"testing"
	"time"

	"github.com/jeckbjy/gsk/arpc"
)

type echoReq struct {
	Text string `json:"text"`
}

type echoRsp struct {
	Text string `json:"text"`
}

func TestRPC(t *testing.T) {
	name := "echo"
	srv := New(name)
	// register callback
	if err := srv.Register(func(ctx arpc.Context, req *echoReq, rsp *echoRsp) error {
		t.Logf("recv msg, %+v", req.Text)
		rsp.Text = fmt.Sprintf("%s world", req.Text)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	go srv.Run()

	time.Sleep(time.Millisecond * 500)

	// synchronous call
	rsp := &echoRsp{}
	_ = srv.Call(name, &echoReq{Text: "sync,hello"}, rsp)
	t.Log(rsp.Text)

	// asynchronous call
	_ = srv.Call(name, &echoReq{Text: "async,hello"}, func(rsp *echoRsp) {
		t.Log(rsp.Text)
	})

	t.Log("wait stop")
	time.Sleep(time.Second * 2)
}
