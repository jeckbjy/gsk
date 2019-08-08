package buffer

import "encoding/binary"

type Reader struct {
	buff *Buffer
}

func (r *Reader) Init(b *Buffer) {
	r.buff = b
}

func (r *Reader) ReadVarintLen() (int, error) {
	size, err := binary.ReadUvarint(r.buff)
	if err != nil {
		return 0, err
	}

	return int(size), nil
}

func (r *Reader) ReadUVarint() (uint, error) {
	size, err := binary.ReadUvarint(r.buff)
	if err != nil {
		return 0, err
	}

	return uint(size), nil
}

func (r *Reader) ReadString(l int) (string, error) {
	s := make([]byte, l, l)
	if n, err := r.buff.Read(s); err != nil || n < l {
		return "", err
	}

	return string(s), nil
}

func (r *Reader) ReadLenString() (string, error) {
	l, err := r.ReadVarintLen()
	if err != nil {
		return "", err
	}

	if l > 0 {
		return r.ReadString(l)
	}

	return "", nil
}
