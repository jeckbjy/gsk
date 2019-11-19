package packet

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/jeckbjy/gsk/util/buffer"
)

type Reader struct {
	buff *buffer.Buffer
	flag uint64
}

func (r *Reader) ReadStringDirect() (string, error) {
	// read string len
	l, err := binary.ReadUvarint(r.buff)
	if err != nil {
		return "", err
	}

	if l > 0 {
		s := make([]byte, l, l)
		_, err := r.buff.Read(s)
		if err != nil {
			return "", err
		}

		return string(s), nil
	}

	return "", nil
}

func (r *Reader) Init(buff *buffer.Buffer) error {
	data := make([]byte, 2)
	if _, err := buff.Read(data); err != nil {
		return err
	}

	flag := binary.LittleEndian.Uint16(data)
	r.flag = uint64(flag)
	r.buff = buff

	return nil
}

func (r *Reader) HasFlag(mask uint64) bool {
	return (r.flag & mask) != 0
}

func (r *Reader) ReadBool(v *bool, mask uint64) {
	*v = r.HasFlag(mask)
}

func (r *Reader) ReadString(v *string, mask uint64) error {
	if r.HasFlag(mask) {
		s, err := r.ReadStringDirect()
		if err != nil {
			return err
		}
		*v = s
	}

	return nil
}

func (r *Reader) ReadMap(m *map[string]string, mask uint64) error {
	if r.HasFlag(mask) {
		l, err := binary.ReadUvarint(r.buff)
		if err != nil {
			return err
		}

		for i := 0; i < int(l); i++ {
			s, err := r.ReadStringDirect()
			if err != nil {
				return err
			}
			index := strings.IndexByte(s, '|')
			if index == -1 {
				return fmt.Errorf("bad map format,%+v", s)
			}
			key := s[:index]
			val := s[index+1:]
			(*m)[key] = val
		}
	}

	return nil
}

func (r *Reader) ReadInt(v *int, mask uint64) error {
	if r.HasFlag(mask) {
		u, err := binary.ReadVarint(r.buff)
		if err != nil {
			return err
		}
		*v = int(u)
	}

	return nil
}

func (r *Reader) ReadInt64(v *int64, mask uint64) error {
	if r.HasFlag(mask) {
		u, err := binary.ReadVarint(r.buff)
		if err != nil {
			return err
		}

		*v = u
	}

	return nil
}

func (r *Reader) ReadUint(v *uint, mask uint64) error {
	if r.HasFlag(mask) {
		u, err := binary.ReadUvarint(r.buff)
		if err != nil {
			return err
		}
		*v = uint(u)
	}

	return nil
}

func (r *Reader) ReadUint64(v *uint64, mask uint64) error {
	if r.HasFlag(mask) {
		u, err := binary.ReadUvarint(r.buff)
		if err != nil {
			return err
		}
		*v = u
	}

	return nil
}

func (r *Reader) ReadUint16(v *uint16, mask uint64) error {
	if r.HasFlag(mask) {
		u, err := binary.ReadUvarint(r.buff)
		if err != nil {
			return err
		}
		if u >= math.MaxUint16 {
			return fmt.Errorf("overflow,%+v", u)
		}

		*v = uint16(u)
	}

	return nil
}
