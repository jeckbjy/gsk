package xor

import (
	"testing"

	"github.com/jeckbjy/gsk/util/buffer"
)

func TestXOR(t *testing.T) {
	//f := New([]byte("asdf"))
	f := xorFilter{key: []byte("asdf")}
	b := buffer.Buffer{}
	b.Append([]byte("12"))
	f.process(&b)
	t.Log(b.String())
	f.process(&b)
	t.Log(b.String())
}
