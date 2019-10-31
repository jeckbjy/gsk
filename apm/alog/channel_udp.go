package alog

import "net"

// UDP协议
type UDPChannel struct {
	BaseChannel
	addr string
	conn net.Conn
}

func (c *UDPChannel) Name() string {
	return "udp"
}

func (c *UDPChannel) SetProperty(key string, value string) error {
	switch key {
	case "addr":
		c.addr = value
	default:
		return c.BaseChannel.SetProperty(key, value)
	}

	return nil
}

func (c *UDPChannel) Open() error {
	if c.conn == nil {
		conn, err := net.Dial("udp", c.addr)
		if err != nil {
			return err
		}
		c.conn = conn
	}

	return nil
}

func (c *UDPChannel) Close() error {
	if c.conn != nil {
		conn := c.conn
		c.conn = nil
		return conn.Close()
	}

	return nil
}

func (c *UDPChannel) Write(msg *Entry) {
	if c.Open() == nil {
		text := msg.Format(c.formatter)
		if text != nil {
			_, _ = c.conn.Write(text)
		}
	}
}
