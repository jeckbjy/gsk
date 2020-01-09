package zset

import (
	"fmt"
	"math/rand"
)

const zMaxLevel = 32

type zCmp func(node *zNode, el *Element, keyEqual bool) bool

type zLevel struct {
	next *zNode
	span uint64
}

type zNode struct {
	key   string
	el    *Element
	prev  *zNode
	level []*zLevel
}

// 一些资料
// http://ju.outofmemory.cn/entry/81525
// http://zhangtielei.com/posts/blog-redis-skiplist.html
type zSkipList struct {
	head  *zNode
	tail  *zNode
	count uint64
	level int
	cmp   zCmp
}

func zNewSkipList(ascending bool) *zSkipList {
	sl := &zSkipList{}
	sl.level = 1
	sl.head = zCreateNode(zMaxLevel, nil)
	sl.count = 0
	sl.cmp = zCmpGreat
	return sl
}

func (sl *zSkipList) SetOrder(ascending bool) {
	if sl.count == 0 {
		if ascending {
			sl.cmp = zCmpGreat
		} else {
			sl.cmp = zCmpLess
		}
	}
}

func (sl *zSkipList) SetDescending() bool {
	if sl.count == 0 {
		sl.cmp = zCmpLess
		return true
	}

	return false
}

func (sl *zSkipList) Clear() {
	sl.level = 1
	sl.count = 0
	sl.tail = nil
	sl.head.prev = nil
	for _, level := range sl.head.level {
		level.next = nil
		level.span = 0
	}
}

// Back返回最后一个元素
func (sl *zSkipList) Back() *Element {
	if sl.tail != nil {
		return sl.tail.el
	}

	return nil
}

func (sl *zSkipList) Len() uint64 {
	return sl.count
}

func (sl *zSkipList) Insert(el *Element) *zNode {
	// update[i] 记录了新数据项的前驱
	update := make([]*zNode, zMaxLevel)
	rank := make([]uint64, zMaxLevel)
	// 查找插入位置，从最高层遍历
	// 同时记录需要更新的节点和排行,用于更新节点和计算span
	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		// store rank that is crossed to reach the insert position
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		//
		for l := x.level[i]; l != nil && l.next != nil && sl.cmp(l.next, el, false); {
			rank[i] += l.span
			x = l.next
			l = x.level[i]
		}
		update[i] = x
	}

	// we assume the element is not already inside, since we allow duplicated
	// scores, reinserting the same element should never happen since the
	// caller of zslInsert() should test in the hash table if the element is
	// already inside or not.
	level := zRandomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			rank[i] = 0
			update[i] = sl.head
			update[i].level[i].span = sl.count
		}
		sl.level = level
	}

	x = zCreateNode(level, el)
	x.key = el.Key
	var lx, lu *zLevel
	for i := 0; i < level; i++ {
		lx = x.level[i]
		lu = update[i].level[i]

		lx.next = lu.next
		lu.next = x

		// update span covered by update[i] as x is inserted here
		lx.span = lu.span - (rank[0] - rank[i])
		lu.span = rank[0] - rank[i] + 1
	}

	// 更高的 level 尚未调整 span
	// increment span for untouched levels
	for i := level; i < sl.level; i++ {
		update[i].level[i].span++
	}

	// 调整新节点的前驱指针
	if update[0] == sl.head {
		x.prev = nil
	} else {
		x.prev = update[0]
	}

	if x.level[0].next != nil {
		x.level[0].next.prev = x
	} else {
		sl.tail = x
	}

	sl.count++
	return x
}

func (sl *zSkipList) DeleteNode(x *zNode, update []*zNode) {
	for i := 0; i < sl.level; i++ {
		lv := update[i].level[i]
		if lv.next == x {
			lv.span += x.level[i].span - 1
			lv.next = x.level[i].next
		} else {
			lv.span--
		}
	}

	if x.level[0].next != nil {
		x.level[0].next.prev = x.prev
	} else {
		sl.tail = x.prev
	}

	for sl.level > 1 && sl.head.level[sl.level-1].next == nil {
		sl.level--
	}
	sl.count--
}

func (sl *zSkipList) Delete(el *Element) bool {
	// find node
	update := make([]*zNode, zMaxLevel)
	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for l := x.level[i]; l.next != nil && sl.cmp(l.next, el, false); {
			x = l.next
			l = x.level[i]
		}
		update[i] = x
	}

	// remove node
	x = x.level[0].next
	if x != nil && x.el.Key == el.Key {
		sl.DeleteNode(x, update)
		return true
	}

	return false
}

// GetRank Find the rank for an element by both score and obj.
// Returns 0 when the element cannot be found, rank otherwise.
// Note that the rank is 1-based due to the span of zsl->head to the
// first element.
func (sl *zSkipList) GetRank(el *Element) uint64 {
	rank := uint64(0)
	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for l := x.level[i]; l.next != nil && sl.cmp(l.next, el, true); {
			rank += x.level[i].span
			x = x.level[i].next
			l = x.level[i]
		}

		// x might be equal to zsl->head, so test if obj is non-NULL
		if x.el != nil && x.el.Key == el.Key {
			return rank
		}
	}

	return 0
}

// Scan all the elements with rank between start and end from the skiplist.
// Start and end are inclusive. Note that start and end need to be 1-based
func (sl *zSkipList) Scan(start, end uint64, cb Callback) {
	var traversed uint64

	// find start
	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].next != nil && (traversed+x.level[i].span) < start {
			traversed += x.level[i].span
			x = x.level[i].next
		}
	}

	traversed++
	x = x.level[0].next
	for x != nil && traversed <= end {
		next := x.level[0].next
		cb(traversed, x.el)
		traversed++
		x = next
	}
}

// Delete all the elements with rank between start and end from the skiplist.
// Start and end are inclusive. Note that start and end need to be 1-based
func (sl *zSkipList) DeleteRangeByRank(start, end uint64, cb Callback) uint64 {
	update := make([]*zNode, zMaxLevel)
	var traversed, removed uint64

	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].next != nil && (traversed+x.level[i].span) < start {
			traversed += x.level[i].span
			x = x.level[i].next
		}
		update[i] = x
	}

	traversed++
	x = x.level[0].next
	for x != nil && traversed <= end {
		next := x.level[0].next
		sl.DeleteNode(x, update)
		cb(traversed, x.el)
		removed++
		traversed++
		x = next
	}

	return removed
}

func (sl *zSkipList) Dump() {
	for i := sl.level - 1; i >= 0; i-- {
		fmt.Printf("lv%+v:\t", i)
		// 忽略head，没有内容
		for n := sl.head; n != nil; n = n.level[i].next {
			l := n.level[i]
			fmt.Printf("[%+v:%v]\t", n.key, l.span)
		}
		fmt.Printf("\n")
	}
}

/////////////////////////////////////////////////////////
// util
/////////////////////////////////////////////////////////
func zCreateNode(level int, el *Element) *zNode {
	n := &zNode{el: el, level: make([]*zLevel, level)}
	for i := range n.level {
		n.level[i] = new(zLevel)
	}

	return n
}

func zRandomLevel() int {
	lv := 1
	for float32(rand.Int31()&0xFFFF) < (0.25 * 0xFFFF) {
		lv++
	}

	if lv < zMaxLevel {
		return lv
	}

	return zMaxLevel

}

func zCmpLess(n *zNode, e *Element, keyEqual bool) bool {
	x := n.el
	if x.Score > e.Score {
		return true
	}

	if x.Score == e.Score {
		if keyEqual {
			return x.Key >= e.Key
		}
		return x.Key > e.Key
	}

	return false
}

func zCmpGreat(n *zNode, e *Element, keyEqual bool) bool {
	x := n.el
	if x.Score < e.Score {
		return true
	}

	if x.Score == e.Score {
		if keyEqual {
			return x.Key <= e.Key
		}

		return x.Key < e.Key
	}

	return false
}
