package alog

type Queue struct {
	head *Entry
	tail *Entry
	len  int
}

func (q *Queue) Empty() bool {
	return q.len == 0
}

func (q *Queue) Len() int {
	return q.len
}

func (q *Queue) Push(e *Entry) {
	if q.head == nil {
		q.head = e
		q.tail = e
	} else {
		q.tail.next = e
		q.tail = e
	}
	q.len++
}

func (q *Queue) Pop() *Entry {
	if q.len == 0 {
		return nil
	}

	result := q.head
	q.len--
	q.head = result.next
	if q.len == 0 {
		q.head = nil
		q.tail = nil
	}

	return result
}

func (q *Queue) Swap(o *Queue) {
	*q, *o = *o, *q
}
