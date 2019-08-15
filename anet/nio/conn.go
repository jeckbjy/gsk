package nio

import (
	"net"

	"github.com/jeckbjy/gsk/anet/nio/internal"
)

type nConn struct {
	selector *internal.Selector
	conn     net.Conn
}

func (c *nConn) Open(conn net.Conn) {

}
