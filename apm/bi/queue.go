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
	n := q.head
	i := 0
	if n != nil {
		for ; i < len(records); i++ {
			records[i] = n
			n = n.next
			records[i].next = nil
			if n == nil {
				break
			}
		}
		if n == nil {
			q.len = 0
			q.head = nil
			q.tail = nil
		} else {
			q.len -= i
			q.head = n
		}
	}

	switch i {
	case 0:
		return nil
	case len(records):
		return records
	default:
		return records[:i]
	}
}
