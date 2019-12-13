package main

import (
	"log"
	"testing"
	"time"

	"github.com/jeckbjy/gsk/anet"
	"github.com/jeckbjy/gsk/anet/tcp"
)

func TestNet(t *testing.T) {
	startServer()
	startClient()
	time.Sleep(time.Second * 1)
	//waitExit()
}

func startServer() {
	log.Printf("start server\n")
	tran := tcp.New()
	tran.AddFilters(
		&LogFilter{},
		&FrameFilter{},
		&HandlerFilter{Peer: "server"})
	_, _ = tran.Listen(":6789")
}

func startClient() {
	log.Printf("start client\n")
	tran := tcp.New()
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

	log.Printf("connect ok:%+v\n", conn)

	_ = conn.Send(&EchoMsg{Text: "ping"})
}

//func waitExit() {
//	fmt.Printf("wait exit\n")
//	sigs := make(chan os.Signal, 1)
//	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
//	<-sigs
//	log.Printf("exit")
//}
