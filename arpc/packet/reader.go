package packet

import (
	"encoding/binary"
	"errors"

	"github.com/jeckbjy/gsk/util/buffer"
)

type Reader struct {
	buff *buffer.Buffer
	flag uint64
}

func (r *Reader) hasFlag(mask uint64) bool {
	return (r.flag & mask) != 0
}

func (r *Reader) Init(buff *buffer.Buffer) error {
	flag, err := binary.ReadUvarint(buff)
	if err != nil {
		return err
	}

	r.flag = flag
	r.buff = buff

	return nil
}

func (r *Reader) ReadBool(v *bool, mask uint64) {
	*v = r.hasFlag(mask)
}

func (r *Reader) ReadString(v *string, mask uint64) error {
	if r.hasFlag(mask) {
		l, err := binary.ReadUvarint(r.buff)
		if err != nil {
			return err
		}
		if l > 0 {
			s := make([]byte, l, l)
			n, err := r.buff.Read(s)
			if err != nil {
				return err
			}
			if n < int(l) {
				return errors.New("no data")
			}
		}

	}

	return nil
}

func (r *Reader) ReadMap(m *map[string]string, mask uint64) error {
	return nil
}

func (r *Reader) ReadInt(v *int, mask uint64) error {
	if r.hasFlag(mask) {
		u, err := binary.ReadVarint(r.buff)
		if err != nil {
			return err
		}
		*v = int(u)
	}

	return nil
}

func (r *Reader) ReadInt64(v *int64, mask uint64) error {
	if r.hasFlag(mask) {
		u, err := binary.ReadVarint(r.buff)
		if err != nil {
			return err
		}

		*v = u
	}

	return nil
}

func (r *Reader) ReadUint(v *uint, mask uint64) error {
	if r.hasFlag(mask) {
		u, err := binary.ReadUvarint(r.buff)
		if err != nil {
			return err
		}
		*v = uint(u)
	}

	return nil
}

func (r *Reader) ReadUint64(v *uint64, mask uint64) error {
	if r.hasFlag(mask) {
		u, err := binary.ReadUvarint(r.buff)
		if err != nil {
			return err
		}
		*v = u
	}

	return nil
}
