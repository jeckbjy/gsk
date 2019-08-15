package internal

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestSelector(t *testing.T) {
	go runServer(t)
	time.Sleep(time.Second)
	go runClient(t)
	time.Sleep(time.Second * 20)
}

func runServer(t *testing.T) {
	log.Printf("start server\n")
	selector, err := New()
	if err != nil {
		t.Log(err)
		return
	}

	l, err := net.Listen("tcp", ":6789")
	if err != nil {
		t.Log(err)
		return
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				break
			}

			if sk, err := selector.Add(conn, OP_READ, nil); err != nil {
				conn.Close()
				t.Log(err)
				break
			} else {
				log.Printf("new connection:%+v\n", sk.FD())
			}
		}
	}()

	for {
		//log.Printf("wait server select\n")
		keys, err := selector.Select()
		if err != nil {
			t.Log(err)
			break
		}

		for _, sk := range keys {
			data := make([]byte, 1024)
			n, err := sk.Read(data)
			if err != nil {
				log.Printf("fail:%+v\n", err)
				continue
			}

			log.Printf("data:%+s\n", data[:n])
			if n, err := sk.Write([]byte("pong")); err != nil || n != 4 {
				log.Printf("send err :%+v, %+v\n", err, n)
			}
		}
	}
}

func runClient(t *testing.T) {
	log.Printf("start client")

	selector, err := New()
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", "localhost:6789")
	if err != nil {
		panic(err)
	}

	log.Printf("dial ok:%+v\n", conn.RemoteAddr().String())

	selector.Add(conn, OP_READ, nil)

	if n, err := conn.Write([]byte("ping")); err == nil {
		log.Printf("send:%+v\n", n)
	}

	for {
		//log.Printf("wait client select\n")
		keys, err := selector.Select()
		if err != nil {
			break
		}

		for _, key := range keys {
			switch {
			case key.Readable():
				bytes := make([]byte, 1024)
				n, err := key.Read(bytes)
				if err != nil {
					log.Printf("read: fd=%+v, err=%+v\n", key.FD(), err)
					continue
				}
				log.Printf("%s:%+v\n", bytes[:n], key.FD())
				n, err = key.Write([]byte("ping"))
				if err != nil || n != 4 {
					log.Printf("write: fd=%+v, err=%+v, %+v\n", key.FD(), err, n)
				}
			}
		}
	}
}
