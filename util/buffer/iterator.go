package buffer

// Iterator 迭代器,只能向前或者向后,不能随意来回迭代
type Iterator struct {
	buff *Buffer
	prev *bnode
	next *bnode
	curr *bnode
}

// 前向迭代
func (iter *Iterator) Prev() bool {
	if iter.prev != nil {
		iter.curr = iter.prev
		iter.prev = iter.prev.prev
		return true
	}

	return false
}

// 正向迭代
func (iter *Iterator) Next() bool {
	if iter.next != nil {
		iter.curr = iter.next
		iter.next = iter.next.next
		return true
	}

	return false
}

// 返回当前数据,调用前必须已经调用过Prev或者Next确认过有数据
func (iter *Iterator) Data() []byte {
	return iter.curr.data
}

// Remove删除当前节点数据,不影响迭代器指针
func (iter *Iterator) Remove() {
	iter.buff.remove(iter.curr)
}
