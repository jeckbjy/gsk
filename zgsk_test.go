package gsk

import (
	"fmt"
	"log"
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
		log.Printf("recv msg, %+v", req.Text)
		rsp.Text = fmt.Sprintf("%s world", req.Text)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	go srv.Run()

	log.Printf("sleep for server start")
	time.Sleep(time.Second)
	log.Printf("try call")

	// synchronous call
	rsp := &echoRsp{}
	if err := srv.Call(name, &echoReq{Text: "sync,hello"}, rsp); err != nil {
		t.Fatal(err)
	} else {
		log.Printf("sync rsp,%s", rsp.Text)
	}

	// asynchronous call
	_ = srv.Call(name, &echoReq{Text: "async,hello"}, func(rsp *echoRsp) {
		log.Printf("async rsp,%s", rsp.Text)
	})

	t.Log("wait stop")
	time.Sleep(time.Second * 5)
	srv.Exit()
	t.Log("finish")
}
