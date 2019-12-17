package main

import (
	"log"
	"testing"
	"time"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/nio"
)

var newTranFunc = nio.New

func TestNet(t *testing.T) {
	startServer()
	startClient()
	time.Sleep(time.Second * 10)
}

func startServer() {
	log.Printf("start server\n")
	tran := newTranFunc()
	tran.AddFilters(
		&LogFilter{},
		&FrameFilter{},
		&HandlerFilter{Peer: "server"})
	_, _ = tran.Listen(":6789")
}

func startClient() {
	log.Printf("start client\n")
	tran := newTranFunc()
	tran.AddFilters(
		&LogFilter{},
		&FrameFilter{},
		&HandlerFilter{Peer: "client"})
	log.Printf("try connect server\n")
	conn, err := tran.Dial("localhost:6789", anet.WithBlocking(true))
	if err != nil {
		log.Printf("dial fail:%+v\n", err)
		return
	}

	log.Printf("connect ok:%+v\n", conn.RemoteAddr())

	_ = conn.Send(&EchoMsg{Text: "ping"})
}
