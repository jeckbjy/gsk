package buffer

import (
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
)

var (
	ErrOverflow     = errors.New("buffer overflow")
	ErrNoEnoughData = errors.New("no enough data")
)

const (
	SeekStart   = io.SeekStart
	SeekCurrent = io.SeekCurrent
	SeekEnd     = io.SeekEnd
)

// New 新建Buffer
func New() *Buffer {
	return &Buffer{}
}

// Swap 交换数据
func Swap(a *Buffer, b *Buffer) {
	*a, *b = *b, *a
}

// Buffer 使用链表的方式管理数据
type Buffer struct {
	blist        // 链表
	len   int    // buffer总长度
	pos   int    // buffer当前位置
	node  *bnode // 当前pos对应的节点,pos=0时可能为空
	off   int    // 相对当前node偏移
	mark  int    // 任意位置标识,不做任何校验
}

// LockedBuffer,增加Mutex,但需要外部手动加锁和释放锁
type LockedBuffer struct {
	sync.Mutex
	Buffer
}

func (b *Buffer) Eof() bool {
	return b.pos >= b.len
}

func (b *Buffer) Empty() bool {
	return b.len == 0
}

// Len return the total length
func (b *Buffer) Len() int {
	return b.len
}

// Pos return the current position
func (b *Buffer) Pos() int {
	return b.pos
}

func (b *Buffer) Mark() int {
	return b.mark
}

func (b *Buffer) SetMark(v int) {
	b.mark = v
}

func (b *Buffer) Bytes() []byte {
	if b.len > 0 {
		b.Concat()
		return b.head.data
	}

	return nil
}

// String 合并并返回string
func (b *Buffer) String() string {
	if b.len > 0 {
		b.Concat()
		return string(b.head.data)
	}

	return ""
}

// 追加其他buffer
func (b *Buffer) AppendBuffer(o *Buffer) {
	for n := o.head; n != nil; n = n.next {
		b.Append(n.data)
	}
}

// Append 追加数据，并将游标指向末尾
func (b *Buffer) Append(data []byte) {
	size := len(data)
	if size > 0 {
		b.pushBack(data)
		b.len += size
		b.pos += size
		b.node = nil
	}
}

// Prepend 在前边添加数据，并修改游标为0
func (b *Buffer) Prepend(data []byte) {
	size := len(data)
	if size > 0 {
		b.pushFront(data)
		b.len += size
		b.pos = 0
		b.off = 0
		b.node = b.head
	}
}

// Peek 读取数据并填充到data中,并返回真实读取的个数
func (b *Buffer) Peek(data []byte) (int, error) {
	var err error
	size := b.len - b.pos
	if len(data) > size {
		data = data[:size]
		err = ErrNoEnoughData
	}

	b.check()
	iter := bfiterator{}
	iter.Read(b, data, false)

	return len(data), err
}

// Read 读取数据并移动当前位置,实现io.Reader接口
func (b *Buffer) Read(data []byte) (int, error) {
	if b.Eof() {
		return 0, io.EOF
	}

	var err error
	size := b.len - b.pos
	if len(data) > size {
		data = data[:size]
		err = ErrNoEnoughData
	}

	b.check()
	iter := bfiterator{}
	iter.Read(b, data, true)
	b.pos += len(data)
	return len(data), err
}

// Write 从当前位置写数据,数据溢出时会自动分配内存, 实现io.Writer接口
// 使用上尽量在外部使用Append创建好数据,Write仅仅提供了一种修改数据的能力,并不希望使用Write来分配内存
func (b *Buffer) Write(data []byte) (int, error) {
	size := len(data)
	overflow := b.pos + size - b.len
	if overflow > 0 {
		// 溢出部分分配空间,是否需要多分配一些内存?
		d := make([]byte, overflow, overflow)
		b.pushBack(d)
		b.len += overflow
	}

	b.check()
	iter := bfiterator{}
	iter.Write(b, data, true)
	b.pos += size
	return size, nil
}

// Seek 移动当前位置,实现io.Seeker接口
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	// 计算当前位置
	var pos int
	switch whence {
	case io.SeekCurrent:
		// offset为负数表示反向移动游标
		pos = b.pos + int(offset)
	case io.SeekStart:
		if offset < 0 {
			return 0, fmt.Errorf("SeekStart offset < 0")
		}
		pos = int(offset)
	case io.SeekEnd:
		if offset < 0 {
			return 0, fmt.Errorf("SeekEnd offset < 0")
		}
		pos = b.len - int(offset)
	}

	if pos < 0 || pos > b.len {
		return 0, ErrOverflow
	}

	if pos == b.pos {
		b.check()
		return int64(b.pos), nil
	}

	switch whence {
	case io.SeekCurrent:
		b.check()
		if offset > 0 { // 向后移动
			iter := bfiterator{}
			iter.Move(b, int(offset), true)
		} else { // 向前移动
			iter := briterator{}
			iter.Move(b, int(-offset), true)
		}
	case io.SeekStart:
		b.node = b.head
		b.off = 0
		iter := bfiterator{}
		iter.Move(b, pos, true)
	case io.SeekEnd:
		b.node = b.tail
		b.off = len(b.node.data)
		iter := briterator{}
		iter.Move(b, int(offset), true)
	}

	b.pos = pos
	return int64(pos), nil
}

// ReadByte 实现接口io.ByteReader
func (b *Buffer) ReadByte() (byte, error) {
	if b.pos >= b.len {
		return 0, ErrOverflow
	}

	b.check()
	// 不存在空数据,但是游标offset可能在末尾
	if b.off >= len(b.node.data) {
		b.node = b.node.next
		b.off = 0
	}

	b.pos++
	b.off++
	return b.node.data[b.off-1], nil
}

// WriteByte 实现接口io.ByteWriter
func (b *Buffer) WriteByte(data byte) error {
	d := [1]byte{}
	d[0] = data
	_, err := b.Write(d[:])
	return err
}

// 发送所有数据
func (b *Buffer) WriteAll(w io.Writer) (int, error) {
	_, _ = b.Seek(0, io.SeekStart)
	size := 0
	for iter := b.head; iter != nil; iter = iter.next {
		n, err := w.Write(iter.data)
		if err != nil {
			return size, err
		}
		size += n
	}

	return size, nil
}

// Clear 清空所有数据
func (b *Buffer) Clear() {
	b.free()
	b.len = 0
	b.pos = 0
	b.off = 0
	b.node = nil
	b.mark = 0
}

// Discard 抛弃当前位置之前的所有数据
func (b *Buffer) Discard() {
	if b.pos == 0 {
		return
	}

	if b.pos == b.len {
		b.Clear()
		return
	}

	left := b.pos
	for left > 0 {
		n := b.head
		size := len(n.data)
		if size > left {
			n.data = n.data[left:]
			break
		}

		left -= size
		b.head = n.next
		n.free()
	}

	b.len -= b.pos
	b.pos = 0
	b.off = 0
	b.node = b.head
}

// Split 从当前位置分隔成两部分
func (b *Buffer) Split() *Buffer {
	if b.pos == 0 {
		return nil
	}

	if b.len == b.pos {
		// 全部都是
		r := New()
		*r = *b
		b.head = nil
		b.Clear()
		return r
	}

	// 部分分割
	left := b.pos
	nlen := 0
	for n := b.head; n != nil; n = n.next {
		nlen++
		size := len(n.data)
		if size < left {
			left -= size
			continue
		}

		if size > left {
			// 分裂成两个
			d := n.data[left:]
			n.data = n.data[:left]
			b.insertBack(n, d)
		}

		r := New()
		r.head = b.head
		r.tail = n
		r.leng = nlen
		r.len = b.pos
		r.pos = b.pos
		r.node = n
		r.off = len(n.data)

		b.head = n.next
		b.leng -= nlen
		b.len -= b.pos
		b.pos = 0
		b.node = b.head
		b.off = 0

		// split
		n.next.prev = nil
		n.next = nil

		return r
	}

	return nil
}

// Concat 合并成一块内存
func (b *Buffer) Concat() {
	if b.leng <= 1 {
		return
	}

	// 拷贝所有节点数据,并保留头部节点,删除其他节点
	data := make([]byte, 0, b.len)
	for n := b.head; n != nil; {
		data = append(data, n.data...)
		t := n
		n = n.next
		t.free()
	}

	b.head.data = data
	b.tail = b.head
	b.node = b.head
	b.off = b.pos
}

// Visit 遍历所有数据
func (b *Buffer) Visit(cb func([]byte) bool) {
	for iter := b.head; iter != nil; iter = iter.next {
		if len(iter.data) != 0 {
			if !cb(iter.data) {
				break
			}
		}
	}
}

func (b *Buffer) Iter() Iterator {
	iter := Iterator{buff: b, prev: b.tail, next: b.head}
	return iter
}

// check 检测一下当前游标是否有效,如果无效则移动到正确位置
func (b *Buffer) check() {
	if b.node != nil || b.len == 0 {
		return
	}

	switch {
	case b.pos == 0:
		b.node = b.head
		b.off = 0
	case b.pos == b.len:
		b.node = b.tail
		b.off = len(b.node.data)
	default:
		left := b.pos
		for n := b.head; n != nil; n = n.next {
			size := len(n.data)
			if size > left {
				b.node = n
				b.off = left
				break
			} else if size == left {
				b.node = n.next
				b.off = 0
			} else {
				left -= size
			}
		}
	}
}

func (b *Buffer) remove(n *bnode) {
	b.len -= len(n.data)
	b.unlink(n)
	if b.pos > b.len {
		b.node = nil
		b.pos = 0
		b.off = 0
	}
}

//////////////////////////////////////////////
// buffer list
//////////////////////////////////////////////
// 链表节点
type bnode struct {
	prev *bnode
	next *bnode
	data []byte
}

func (n *bnode) free() {
	n.prev = nil
	n.next = nil
	n.data = nil
}

// 双向非循环链表
type blist struct {
	head *bnode // 链表头
	tail *bnode // 链表尾
	leng int    // node length
}

func (l *blist) pushBack(data []byte) {
	n := &bnode{}
	n.data = data

	if l.tail != nil {
		t := l.tail
		t.next = n
		n.prev = t

		l.tail = n
	} else {
		l.head = n
		l.tail = n
	}

	l.leng++
}

func (l *blist) pushFront(data []byte) {
	n := &bnode{}
	n.data = data

	if l.head != nil {
		h := l.head
		h.prev = n
		n.next = h
		l.head = n
	} else {
		l.head = n
		l.tail = n
	}

	l.leng++
}

func (l *blist) insertBack(p *bnode, data []byte) {
	nx := p.next

	n := &bnode{}
	n.data = data

	n.prev = p
	n.next = p.next
	p.next = n
	if nx != nil {
		nx.prev = n
	} else {
		l.tail = n
	}

	l.leng++
}

// 删除节点
func (l *blist) unlink(n *bnode) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}

	n.prev = nil
	n.next = nil
	n.data = nil
	l.leng--
}

// 释放所有节点
func (l *blist) free() {
	for n := l.head; n != nil; {
		t := n
		n = n.next
		t.free()
	}

	l.head = nil
	l.tail = nil
	l.leng = 0
}

// bnode pool
//type bpool struct {
//}

//////////////////////////////////////////////
// buffer forward iterator
//////////////////////////////////////////////
type bfiterator struct {
	node *bnode // 当前节点
	offs int    // 当前偏移
	size int    // 需要读取的总长度
	curr int    // 当前处理的数据
	last int    // 上次处理的数据
	data []byte // 数据
}

func (fi *bfiterator) Init(node *bnode, offs int, size int) {
	if size == -1 { // -1表示处理到最后
		size = math.MaxInt32
	}

	fi.node = node
	fi.offs = offs
	fi.size = size
	fi.curr = 0
	fi.last = 0
}

// Next 返回是否还有数据可以处理
func (fi *bfiterator) Next() bool {
	if fi.node == nil || fi.curr >= fi.size {
		return false
	}

	fi.last = fi.curr
	need := fi.size - fi.curr
	for {
		data := fi.node.data
		left := len(data) - fi.offs
		if left <= 0 {
			// ignore zero data
			fi.node = fi.node.next
			fi.offs = 0
			if fi.node == nil {
				return false
			}
			continue
		}

		// read data
		if left > need {
			fi.data = data[fi.offs : fi.offs+need]
			fi.offs += need
			fi.curr += need
		} else {
			fi.data = data[fi.offs:]
			fi.node = fi.node.next
			fi.offs = 0
			fi.curr += left
		}
		break
	}

	return true
}

// Read 从buffer中读取数据,保存到data中
// back 是否回写buffer游标
func (fi *bfiterator) Read(b *Buffer, data []byte, back bool) {
	fi.Init(b.node, b.off, len(data))
	for fi.Next() {
		copy(data[fi.last:], fi.data)
	}

	if back {
		b.node = fi.node
		b.off = fi.offs
	}
}

// Write 将data中数据写入到buffer中
// back 是否回写buffer游标
func (fi *bfiterator) Write(b *Buffer, data []byte, back bool) {
	fi.Init(b.node, b.off, len(data))
	for fi.Next() {
		copy(fi.data, data[fi.last:])
	}

	if back {
		b.node = fi.node
		b.off = fi.offs
	}
}

// Move 从当前位置向后移动游标，移动size个位置
func (fi *bfiterator) Move(b *Buffer, size int, back bool) {
	fi.Init(b.node, b.off, size)
	left := fi.size
	for left > 0 && fi.node != nil {
		tail := len(fi.node.data) - fi.offs
		if tail <= left {
			left -= tail
			fi.node = fi.node.next
			fi.offs = 0
		} else {
			fi.offs += left
			//left = 0
			break
		}
	}

	if back {
		b.node = fi.node
		b.off = fi.offs
	}
}

//////////////////////////////////////////////
// buffer reverse iterator
//////////////////////////////////////////////
type briterator struct {
	node *bnode // 当前节点
	offs int    // 当前偏移
	size int    // 需要读取的总长度
	curr int    // 当前处理的数据
	last int    // 上次处理的数据
	data []byte // 数据
}

func (ri *briterator) Init(node *bnode, offset int, size int) {
	if offset == -1 {
		offset = len(node.data)
	}
	ri.node = node
	ri.offs = offset
	ri.size = size
	ri.curr = 0
	ri.last = 0
}

func (ri *briterator) Next() bool {
	if ri.node == nil || ri.curr >= ri.size {
		return false
	}

	ri.last = ri.curr
	need := ri.size - ri.curr
	for {
		if ri.offs == 0 {
			ri.node = ri.node.next
			ri.offs = 0
			if ri.node == nil {
				return false
			}

			continue
		}

		data := ri.node.data
		if ri.offs == -1 {
			ri.offs = len(data)
		}

		if ri.offs > need {
			ri.data = data[ri.offs-need : ri.offs]
			ri.curr += need
		} else {
			ri.data = data[:ri.offs]
			ri.node = ri.node.prev
			ri.offs = -1
			ri.curr += ri.offs
		}
		break
	}

	return true
}

// Move 向前移动size个字节
func (ri *briterator) Move(b *Buffer, size int, back bool) {
	ri.Init(b.node, b.off, size)
	left := ri.size
	for left > 0 && b.node != nil {
		if ri.offs < left {
			left -= ri.offs

			ri.node = ri.node.prev
			ri.offs = len(ri.node.data)
		} else {
			ri.offs -= left
			break
		}
	}

	if back {
		b.node = ri.node
		b.off = ri.offs
	}
}
