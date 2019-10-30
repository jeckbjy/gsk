package basex

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

var (
	ErrNotSupport = errors.New("not support, string len must be 32 or 64")
)

const (
	Default = "0123456789bcdefghjkmnpqrstuvwxyz"
)

var encoding, _ = NewEncoding(Default)

func SetDefault(e *Encoding) {
	target := (*unsafe.Pointer)(unsafe.Pointer(&encoding))
	source := unsafe.Pointer(e)
	atomic.SwapPointer(target, source)
}

func Encode(x uint64) string {
	return encoding.Encode(x)
}

func Decode(s string) uint64 {
	return encoding.Decode(s)
}

const invalid = 0xff

func NewEncoding(abc string) (*Encoding, error) {
	e := &Encoding{}
	if err := e.init(abc); err != nil {
		return nil, err
	}

	return e, nil
}

// base32或者base64
// 将uint64转换为string
// https://github.com/mmcloughlin/geohash/blob/master/base32.go
type Encoding struct {
	encode string
	decode [128]byte
}

func (e *Encoding) init(s string) error {
	l := len(s)
	if l != 32 && l != 64 {
		return ErrNotSupport
	}

	e.encode = s
	for i := 0; i < len(e.decode); i++ {
		e.decode[i] = invalid
	}

	for i := 0; i < len(s); i++ {
		e.decode[s[i]] = byte(i)
	}

	return nil
}

func (e *Encoding) Encode(x uint64) string {
	if len(e.encode) == 32 {
		b := [13]byte{}
		for i := 12; i >= 0; i-- {
			b[i] = e.encode[x&0x1f]
			x >>= 5
			if x == 0 {
				return string(b[i:])
			}
		}

		return string(b[:])
	} else {
		b := [11]byte{}
		for i := 10; i >= 0; i-- {
			b[i] = e.encode[x&0x3f]
			x >>= 6
			if x == 0 {
				return string(b[i:])
			}
		}
		return string(b[:])
	}
}

func (e *Encoding) Decode(s string) uint64 {
	if len(e.encode) == 32 {
		x := uint64(0)
		for i := 0; i < len(s); i++ {
			x = (x << 5) | uint64(e.decode[s[i]])
		}

		return x
	} else {
		x := uint64(0)
		for i := 0; i < len(s); i++ {
			x = (x << 6) | uint64(e.decode[s[i]])
		}
		return x
	}
}
