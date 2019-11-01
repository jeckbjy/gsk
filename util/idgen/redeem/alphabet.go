package redeem

import (
	"bytes"
	"errors"
)

var (
	ErrBadData = errors.New("bad data")
)

const DefaultABC = "ABCDEFGHJKLMNPQRSTUVWXYZ"

const invalid = -1

// NewAlphabet creates a new alphabet from the passed string.
func NewAlphabet(abc string, seed uint64) *Alphabet {
	ret := new(Alphabet)
	ret.encode = shuffle(abc, seed)
	for i := range ret.decode {
		ret.decode[i] = (invalid)
	}
	for i, b := range ret.encode {
		ret.decode[b] = int8(i)
	}

	return ret
}

func shuffle(alphabet string, seed uint64) string {
	source := []rune(alphabet)
	length := len(source)
	for i := length - 1; i >= 0; i-- {
		seed = (seed*9301 + 49297) % 233280
		j := int(seed * uint64(length) / 233280)
		source[i], source[j] = source[j], source[i]
	}

	return string(source)
}

func reverse(r []byte) string {
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return string(r)
}

type Alphabet struct {
	encode string
	decode [128]int8
}

func (a *Alphabet) Encode(x uint64) string {
	// base n 编码
	buff := bytes.Buffer{}
	leng := uint64(len(a.encode))
	for x > 0 {
		i := x % leng
		x /= leng
		b := a.encode[i]
		buff.WriteByte(b)
	}

	return reverse([]byte(buff.String()))
}

func (a *Alphabet) Decode(s string) (uint64, error) {
	x := uint64(0)
	leng := uint64(len(a.encode))

	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 128 || int(b) == invalid {
			return 0, ErrBadData
		}
		n := a.decode[b]
		x = x*leng + uint64(n)
	}

	return x, nil
}
