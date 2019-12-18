package internal

import (
	"log"
	"syscall"
	"testing"
	"time"
)

func doAccept(p Poller, l *Listener, event *Event) {
	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("accept,%+v", conn.Fd())
	if err := p.Add(conn.Fd()); err != nil {
		log.Fatal(err)
	}
}

func doRead(event *Event) {
	// 这里没有做粘包处理,假定都能立即读取和发送
	data := make([]byte, 1024)
	n, err := syscall.Read(event.fd, data)
	if err != nil {
		log.Fatal(err)
	}

	str := string(data[:n])
	log.Print(str)
	result := "pong"
	if str == "pong" {
		result = "ping"
	}
	_, err1 := syscall.Write(event.fd, []byte(result))
	if err1 != nil {
		log.Fatal(err1)
	}
}

func TestPoll(t *testing.T) {
	poller := newPoller()
	if err := poller.Open(); err != nil {
		t.Fatal(err)
	}

	// start server
	listener, err := Listen("tcp", ":6789")
	if err != nil {
		t.Fatal(err)
	}

	_ = poller.Add(listener.Fd())

	// start Poller
	go func() {
		for {
			err := poller.Wait(func(event *Event) {
				if event.events&EventRead != 0 {
					if event.Fd() == listener.Fd() {
						doAccept(poller, listener, event)
					} else {
						doRead(event)
					}
				}
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	// start client
	conn, err := Dial("tcp", "localhost:6789")
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("local %+v, remote %+v", conn.LocalAddr().String(), conn.RemoteAddr().String())

	_ = poller.Add(conn.Fd())
	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 1)
}
