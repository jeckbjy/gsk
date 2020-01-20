package base

import (
	"github.com/jeckbjy/gsk/exec"
)

type Node struct {
	prev *Node
	next *Node
	task exec.Task
}

// 双向非循环队列
type Queue struct {
	head *Node
	tail *Node
}

func (q *Queue) Swap(o *Queue) {
	*q, *o = *o, *q
}

func (q *Queue) Empty() bool {
	return q.head == nil
}

func (q *Queue) Push(task exec.Task) {
	n := gPool.Obtain()
	n.task = task
	q.pushNode(n)
}

func (q *Queue) Pop() exec.Task {
	n := q.popNode()
	if n != nil {
		t := n.task
		gPool.Free(n)
		return t
	}

	return nil
}

func (q *Queue) pushNode(n *Node) {
	if q.tail == nil {
		q.head = n
		q.tail = n
	} else {
		n.prev = q.tail
		q.tail.next = n
		q.tail = n
	}
}

// 获取第一个node
func (q *Queue) popNode() *Node {
	if q.head != nil {
		node := q.head
		next := node.next
		q.head = next
		if next == nil {
			q.tail = nil
		}

		node.next = nil
		return node
	}

	return nil
}

func (q *Queue) Process() {
	for !q.Empty() {
		task := q.Pop()
		_ = task.Run()
	}
}
