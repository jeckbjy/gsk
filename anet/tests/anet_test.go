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
	listener := startServer()
	conn := startClient()
	time.Sleep(time.Second * 10)
	_ = listener.Close()
	_ = conn.Close()
	log.Printf("finish")
}

func startServer() anet.Listener {
	log.Printf("start server\n")
	tran := newTranFunc()
	tran.AddFilters(
		&LogFilter{},
		&FrameFilter{},
		&HandlerFilter{Peer: "server"})
	l, err := tran.Listen(":6789")
	if err != nil {
		log.Fatal(err)
	}
	return l
}

func startClient() anet.Conn {
	log.Printf("start client\n")
	tran := newTranFunc()
	tran.AddFilters(
		&LogFilter{},
		&FrameFilter{},
		&HandlerFilter{Peer: "client"})
	log.Printf("try connect server\n")
	conn, err := tran.Dial("localhost:6789", anet.WithBlocking(true))
	if err != nil {
		log.Fatalf("dial fail:%+v\n", err)
	}

	log.Printf("connect ok:%+v\n", conn.RemoteAddr())

	_ = conn.Send(&EchoMsg{Text: "ping"})
	return conn
}
