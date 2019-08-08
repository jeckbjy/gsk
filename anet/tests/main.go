package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeckbjy/micro/anet"
	_ "github.com/jeckbjy/micro/anet/tcp"
	"github.com/jeckbjy/micro/util/log"
)

func main() {
	startServer()
	startClient()
	waitExit()
}

type EchoMsg struct {
	Text string
}

func startServer() {
	fmt.Printf("start server\n")
	tran := anet.NewTran()
	tran.AddFilters(
		&LogFilter{},
		&FrameFilter{},
		&HandlerFilter{Peer: "server"})
	_, _ = tran.Listen(":6789")
}

func startClient() {
	fmt.Printf("start client\n")
	tran := anet.NewTran()
	tran.AddFilters(
		&LogFilter{},
		&FrameFilter{},
		&HandlerFilter{Peer: "client"})
	fmt.Printf("try connect server\n")
	conn, err := tran.Dial("localhost:6789", anet.WithBlocking(true))
	if err != nil {
		fmt.Printf("dial fail:%+v\n", err)
		return
	}

	fmt.Printf("connect ok:%+v\n", conn)

	_ = conn.Send(&EchoMsg{Text: "ping"})
}

func waitExit() {
	fmt.Printf("wait exit\n")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Log("exit")
}
