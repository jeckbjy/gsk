package internal

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestSelector(t *testing.T) {
	quit := false
	selector, err := New()
	if err != nil {
		t.Fatal(err)
	}

	// start poller
	go func() {
		for !quit {
			err := selector.Wait(func(ev *Event) {
				if ev.Readable() {
					// 粘包处理，写监听处理，socket关闭处理
					data := make([]byte, 128)
					n, err := ev.Read(data)
					if err != nil {
						return
					}

					str := string(data[:n])
					log.Printf("%+v", str)
					if str == "ping" {
						_, _ = ev.Write([]byte("pong"))
					} else {
						_, _ = ev.Write([]byte("ping"))
					}
				}
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	// start server
	l, err := net.Listen("tcp", ":6789")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for !quit {
			conn, err := l.Accept()
			if err != nil {
				break
			}
			if err := selector.Add(conn); err != nil {
				t.Fatal(err)
			}
		}
	}()

	// start client
	client, err := net.Dial("tcp", "localhost:6789")
	if err != nil {
		t.Fatal(err)
	}
	_ = selector.Add(client)
	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 200)
	quit = true
	_ = l.Close()
	_ = client.Close()
	_ = selector.Wakeup()
}
