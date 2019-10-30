package packet

import (
	"fmt"

	"github.com/jeckbjy/gsk/util/buffer"
)

type Writer struct {
	buffer.Writer
	flag uint64
}

func (w *Writer) WriteBool(v bool, mask uint64) {
	w.flag |= mask
}

func (w *Writer) WriteString(v string, mask uint64) {
	if len(v) > 0 {
		w.flag |= mask
		w.PutLenString(v)
	}
}

func (w *Writer) WriteMap(m map[string]string, mask uint64) {
	if len(m) > 0 {
		w.flag |= mask
		w.PutVarintLen(len(m))
		for k, v := range m {
			s := fmt.Sprintf("%s:%s", k, v)
			w.PutLenString(s)
		}
	}
}

func (w *Writer) WriteInt(v int, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutVarint(v)
	}
}

func (w *Writer) WriteInt32(v int32, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutVarint32(v)
	}
}

func (w *Writer) WriteInt64(v int64, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutVarint64(v)
	}
}

func (w *Writer) WriteUint(v uint, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutUVarint(v)
	}
}

func (w *Writer) WriteUint32(v uint32, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutUVarint32(v)
	}
}

func (w *Writer) WriteUint64(v uint64, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutUVarint64(v)
	}
}
