package base

import "github.com/jeckbjy/gsk/exec"

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

func (l *Queue) Swap(o *Queue) {
	*l, *o = *o, *l
}

func (l *Queue) Empty() bool {
	return l.head == nil
}

func (l *Queue) Push(task exec.Task) {
	n := gPool.Obtain()
	n.task = task
	if l.tail == nil {
		l.head = n
		l.tail = n
	} else {
		n.prev = l.tail
		l.tail.next = n
	}
}

func (l *Queue) Pop() exec.Task {
	if l.head != nil {
		n := l.head
		l.head = n.next
		n.next = nil
		n.prev = nil
		t := n.task
		gPool.Free(n)
		return t
	}

	return nil
}

func (l *Queue) Free(n *Node) {
	gPool.Free(n)
}

func (l *Queue) Process() {
	for !l.Empty() {
		task := l.Pop()
		task.Run()
	}
}
