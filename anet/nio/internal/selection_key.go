package internal

type SelectionKey struct {
	sock      interface{} // net.Conn or net.Listener
	data      interface{} // bind data
	fd        uintptr     // file descriptor
	interests Operation   // registered ops
	ready     Operation   // ready ops
}

func (s *SelectionKey) reset() {
	s.ready = 0
}

func (s *SelectionKey) FD() uintptr {
	return s.fd
}

func (s *SelectionKey) Readable() bool {
	return s.ready&OP_READ != 0
}

func (s *SelectionKey) Writable() bool {
	return s.ready&OP_WRITE != 0
}

func (s *SelectionKey) setReadable() {
	if s.interests&OP_READ != 0 {
		s.ready |= OP_READ
	}
}

func (s *SelectionKey) setWritable() {
	if s.interests&OP_WRITE != 0 {
		s.ready |= OP_WRITE
	}
}

func (sk *SelectionKey) Read(b []byte) (int, error) {
	return Read(sk.fd, b)
}

func (sk *SelectionKey) Write(b []byte) (int, error) {
	return Write(sk.fd, b)
}
