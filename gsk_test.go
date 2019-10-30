package gsk

import (
	"fmt"
	"testing"
)

type echoReq struct {
	Text string `json:"text"`
}

type echoRsp struct {
	Text string `json:"text"`
}

func TestARPC(t *testing.T) {
	name := "echo"
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}

	// 注册消息ID??

	// 注册消息回调
	s.Register(func(req *echoReq, rsp *echoRsp) {
		t.Logf("recv msg,%+v", req.Text)
		rsp.Text = fmt.Sprintf("%s world", req.Text)
	})

	// 运行服务器
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	// client
	// 同步调用
	rsp := &echoRsp{}
	_ = s.Call(name, &echoReq{Text: "sync, hello"}, rsp)
	t.Log("wait sync reply")
	t.Log(rsp.Text)

	// 异步调用
	s.Call(name, &echoReq{Text: "ping async"}, func(rsp *echoRsp) {
		t.Logf("reply:%s", rsp.Text)
	})
	t.Log("wait async reply")

	_ = s.Stop()
}
