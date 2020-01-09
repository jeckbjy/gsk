package timingwheel

func newBucket() *bucket {
	return &bucket{}
}

// 双向非循环列表
type bucket struct {
	head *Timer
	tail *Timer
	size int
}

func (b *bucket) Reset() {
	b.head = nil
	b.tail = nil
	b.size = 0
}

func (b *bucket) Empty() bool {
	return b.size == 0
}

func (b *bucket) Len() int {
	return b.size
}

func (b *bucket) Front() *Timer {
	return b.head
}

func (b *bucket) Push(t *Timer) {
	if b.tail != nil {
		b.tail.next = t
		b.tail = t
	} else {
		b.head = t
		b.tail = t
	}
	b.size++
	t.list = b
}

func (b *bucket) Remove(t *Timer) {
	if t.prev != nil {
		t.prev.next = t.next
	}
	if t.next != nil {
		t.next.prev = t.prev
	}

	t.list = nil
	t.prev = nil
	t.next = nil
	b.size--
}

// 注意:为了减少遍历,这里timer所关联的list并没有同步修改,在外部特定个地方一次性处理
func (b *bucket) merge(other *bucket) {
	if b.tail != nil {
		b.tail.next = other.head
		b.tail = other.tail
		b.size += other.size
	} else {
		*b = *other
	}

	other.Reset()
}

// 取消关联
func (b *bucket) unlink() {
	for iter := b.head; iter != nil; iter = iter.next {
		iter.list = nil
	}
}
