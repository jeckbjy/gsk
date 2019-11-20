package bi

// 单向非循环队列
type Queue struct {
	head *Message
	tail *Message
	len  int
}

func (q *Queue) Empty() bool {
	return q.len == 0
}

func (q *Queue) Len() int {
	return q.len
}

func (q *Queue) Push(m *Message) {
	if q.head == nil {
		q.head = m
		q.tail = m
	} else {
		q.tail.next = m
		q.tail = m
	}
	q.len++
}

// 返回N个数据
func (q *Queue) Pop(records []*Message) []*Message {
	if q.head == nil {
		return nil
	}

	node := q.head
	idx := 0
	for ; idx < len(records) && node != nil; idx++ {
		tmp := node
		node = node.next
		tmp.next = nil
		records[idx] = tmp
	}

	if node == nil {
		q.len = 0
		q.head = nil
		q.tail = nil
	} else {
		q.len -= idx
		q.head = node
	}

	if idx == len(records) {
		return records
	} else {
		return records[:idx]
	}
}
