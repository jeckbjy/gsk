package packet

import (
	"encoding/binary"
	"fmt"
)

// 用于序列化Head信息
// 两个字节用于标识flag
// 编码上整数都是用Varint编码
// String是用len+str形式编码
type Writer struct {
	data []byte
	flag uint64
}

func (w *Writer) Init() {
	w.data = make([]byte, 2, 128)
}

func (w *Writer) Flush() []byte {
	f := uint16(w.flag)
	d := make([]byte, 2)
	binary.LittleEndian.PutUint16(d, f)
	w.data[0] = d[0]
	w.data[1] = d[1]
	return w.data
}

func (w *Writer) PutLen(l int) {
	w.PutUvarint(uint64(l))
}

func (w *Writer) PutVarint(i int64) {
	d := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(d, i)
	w.PutBytes(d[:n])
}

func (w *Writer) PutUvarint(u uint64) {
	d := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(d, u)
	w.PutBytes(d[:n])
}

func (w *Writer) PutString(s string) {
	if len(s) > 0 {
		w.PutLen(len(s))
		w.PutBytes([]byte(s))
	}
}

func (w *Writer) PutBytes(v []byte) {
	w.data = append(w.data, v...)
}

func (w *Writer) WriteBool(v bool, mask uint64) {
	w.flag |= mask
}

func (w *Writer) WriteString(v string, mask uint64) {
	if len(v) > 0 {
		w.flag |= mask
		w.PutString(v)
	}
}

func (w *Writer) WriteMap(m map[string]string, mask uint64) {
	if len(m) > 0 {
		w.flag |= mask
		w.PutLen(len(m))
		for k, v := range m {
			s := fmt.Sprintf("%s|%s", k, v)
			w.PutString(s)
		}
	}
}

func (w *Writer) WriteInt(v int, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutVarint(int64(v))
	}
}

func (w *Writer) WriteInt64(v int64, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutVarint(v)
	}
}

func (w *Writer) WriteUint(v uint, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutUvarint(uint64(v))
	}
}

func (w *Writer) WriteUint64(v uint64, mask uint64) {
	if v != 0 {
		w.flag |= mask
		w.PutUvarint(v)
	}
}
