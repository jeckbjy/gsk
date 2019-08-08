package buffer

import "encoding/binary"

// Writer 带有缓存的写Buffer
type Writer struct {
	buff  *Buffer // 最终数据
	data  []byte  // 数据缓存
	size  int     // 当前大小
	chunk int     // 每次分配大小
}

func (w *Writer) Init(b *Buffer, chunk int) {
	w.buff = b
	w.chunk = chunk
}

// obtain 预先获取数据
func (w *Writer) obtain(need int) []byte {
	if need > w.chunk-w.size {
		if w.size > 0 {
			w.buff.Append(w.data[:w.size])
		}
		w.data = make([]byte, w.chunk, w.chunk)
		w.size = 0
	}

	return w.data[w.size:]
}

// move 增加size
func (w *Writer) move(n int) {
	w.size += n
}

func (w *Writer) space() int {
	return w.chunk - w.size
}

func (w *Writer) Flush() {
	if w.size > 0 {
		w.buff.Append(w.data[:w.size])
		w.data = nil
		w.size = 0
	}
}

func (w *Writer) PutVarintLen(i int) {
	w.PutUVarint64(uint64(i))
}

func (w *Writer) PutVarint(i int) {
	w.PutVarint64(int64(i))
}

func (w *Writer) PutUVarint(u uint) {
	w.PutUVarint64(uint64(u))
}

func (w *Writer) PutVarint32(i int32) {
	d := w.obtain(binary.MaxVarintLen32)
	n := binary.PutVarint(d, int64(i))
	w.move(n)
}

func (w *Writer) PutUVarint32(u uint32) {
	d := w.obtain(binary.MaxVarintLen32)
	n := binary.PutUvarint(d, uint64(u))
	w.move(n)
}

func (w *Writer) PutVarint64(i int64) {
	d := w.obtain(binary.MaxVarintLen64)
	n := binary.PutVarint(d, i)
	w.move(n)
}

func (w *Writer) PutUVarint64(u uint64) {
	d := w.obtain(binary.MaxVarintLen64)
	n := binary.PutUvarint(d, u)
	w.move(n)
}

func (w *Writer) PutByte(b byte) {
	d := w.obtain(1)
	d[0] = b
	w.move(1)
}

/**
 * BigEndian
 */
func (w *Writer) PutInt16BE(i int16) {
	d := w.obtain(2)
	binary.BigEndian.PutUint16(d, uint16(i))
	w.move(2)
}

func (w *Writer) PutInt32BE(i int32) {
	d := w.obtain(4)
	binary.BigEndian.PutUint32(d, uint32(i))
	w.move(4)
}

func (w *Writer) PutInt64BE(i int64) {
	d := w.obtain(8)
	binary.BigEndian.PutUint64(d, uint64(i))
	w.move(8)
}

func (w *Writer) PutUInt16BE(i uint16) {
	d := w.obtain(2)
	binary.BigEndian.PutUint16(d, i)
	w.move(2)
}

func (w *Writer) PutUInt32BE(i uint32) {
	d := w.obtain(4)
	binary.BigEndian.PutUint32(d, i)
	w.move(4)
}

func (w *Writer) PutUInt64BE(i uint64) {
	d := w.obtain(8)
	binary.BigEndian.PutUint64(d, i)
	w.move(8)
}

/**
 * LittleEndian
 */

func (w *Writer) PutInt16LE(i int16) {
	d := w.obtain(2)
	binary.LittleEndian.PutUint16(d, uint16(i))
	w.move(2)
}

func (w *Writer) PutInt32LE(i int32) {
	d := w.obtain(4)
	binary.LittleEndian.PutUint32(d, uint32(i))
	w.move(4)
}

func (w *Writer) PutInt64LE(i int64) {
	d := w.obtain(8)
	binary.LittleEndian.PutUint64(d, uint64(i))
	w.move(8)
}

func (w *Writer) PutUInt16LE(i uint16) {
	d := w.obtain(2)
	binary.LittleEndian.PutUint16(d, i)
	w.move(2)
}

func (w *Writer) PutUInt32LE(i uint32) {
	d := w.obtain(4)
	binary.LittleEndian.PutUint32(d, i)
	w.move(4)
}

func (w *Writer) PutUInt64LE(i uint64) {
	d := w.obtain(8)
	binary.LittleEndian.PutUint64(d, i)
	w.move(8)
}

func (w *Writer) PutString(s string) {
	w.PutBytes([]byte(s))
}

func (w *Writer) PutLenString(s string) bool {
	if len(s) > 0 {
		w.PutVarintLen(len(s))
		w.PutString(s)
		return true
	}

	return false
}

func (w *Writer) PutBytes(d []byte) {
	if w.space() >= len(d) {
		// 拷贝
		copy(w.data[w.size:], d)
		w.move(len(d))
	} else {
		w.Flush()
		w.buff.Append(d)
	}
}
