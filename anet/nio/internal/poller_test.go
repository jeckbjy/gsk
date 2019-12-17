package internal

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestPoll(t *testing.T) {
	poller := newPoller()
	if err := poller.Open(); err != nil {
		t.Fatal(err)
	}

	// start poller
	go func() {
		for {
			err := poller.Wait(func(event *Event) {
				if event.events&EventRead != 0 {
					// 这里没有做粘包处理,假定都能立即读取和发送
					data := make([]byte, 1024)
					n, err := Read(event.fd, data)
					if err != nil {
						t.Fatal(err)
					}
					str := string(data[:n])
					log.Print(str)
					result := "pong"
					if str == "pong" {
						result = "ping"
					}
					_, err1 := Write(event.fd, []byte(result))
					if err1 != nil {
						t.Fatal(err1)
					}
				}
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	// start server
	go func() {
		l, err := net.Listen("tcp", ":6789")
		if err != nil {
			t.Fatal(err)
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				t.Fatal(err)
			}
			fd, err := GetFD(conn)
			if err != nil {
				t.Fatal(err)
			}

			if err := SetNonblock(fd); err != nil {
				t.Fatal(err)
			}

			if err := poller.Add(fd); err != nil {
				t.Fatal(err)
			}
		}
	}()

	// start client
	go func() {
		conn, err := net.Dial("tcp", "localhost:6789")
		if err != nil {
			panic(err)
		}
		fd, err := GetFD(conn)
		if err != nil {
			t.Fatal(err)
		}

		if err := SetNonblock(fd); err != nil {
			t.Fatal(err)
		}
		if err := poller.Add(fd); err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Write([]byte("ping")); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Second * 1)
}
