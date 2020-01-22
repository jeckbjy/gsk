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
	// 启动服务器
	name := "echo"
	srv := New(name)
	// 注册回调: register callback
	if err := srv.Register(func(ctx arpc.Context, req *echoReq, rsp *echoRsp) error {
		log.Printf("[server] recv msg,%+v\n", req.Text)
		rsp.Text = fmt.Sprintf("%s world", req.Text)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	go srv.Run()

	log.Printf("wait for server start")
	time.Sleep(time.Millisecond * 20)

	// 同步RPC调用 synchronous call
	log.Printf("[client] try call sync")
	rsp := &echoRsp{}
	if err := srv.Call(name, &echoReq{Text: "sync hello"}, rsp); err != nil {
		t.Fatal(err)
	} else {
		log.Printf("[client] recv response,%s", rsp.Text)
	}

	// 异步RPC调用 asynchronous call
	log.Printf("[client] try call async")
	err := srv.Call(name, &echoReq{Text: "async hello"}, func(rsp *echoRsp) error {
		log.Printf("[client] recv response,%s", rsp.Text)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	} else {
		log.Printf("[client] async call ok")
	}

	time.Sleep(time.Second * 2)
	srv.Exit()
	t.Log("finish")
}
