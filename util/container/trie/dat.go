package trie

import (
	"fmt"
)

// NewDATrie return DoubleArrayTrie
func NewDATrie() *DoubleArrayTrie {
	dat := &DoubleArrayTrie{}
	return dat
}

// DoubleArrayTrie trie
type DoubleArrayTrie struct {
	cells []zCell
}

// Build 构建一个DATrie
// 要求:
//  1:必须utf8编码
//  2:必须不能重复，且递增有序,即必须保证相同前缀的短的在前边,否则会导致访问越界
func (dat *DoubleArrayTrie) Build(words []string) error {
	if len(words) == 0 {
		return nil
	}

	dat.cells = make([]zCell, len(words)*10)

	nodes := newList()
	pools := newList()
	// for cache
	siblings := make([]*zNode, 0, 32)

	// 根节点
	root := &zNode{}
	root.init(0, 0, 0, len(words))
	root.index = 0
	nodes.PushBack(root)

	// 开始搜索的位置
	nextPos := 1

	// 广度优先搜索创建Trie树
	// 查找区间每一层node节点[left,right]之间不同的字符,用于构建Trie下一级索引
	//
	for nodes.Len() != 0 {
		parent := nodes.PopFront()
		pools.PushBack(parent)
		p := *parent

		// step1: find siblings
		siblings = siblings[:0]
		for idx := p.left; idx < p.right; {
			start := idx
			code := words[idx][p.depth]
			// fmt.Printf("%c\n", code)

			// find end index
			for idx++; idx < p.right; idx++ {
				// 保证错误数据不会导致越界
				if p.depth > len(words[idx]) {
					return fmt.Errorf("words must order asending:%+v", words[idx])
				}

				if words[idx][p.depth] != code {
					break
				}
			}

			// new node
			var node *zNode
			if pools.Len() > 0 {
				node = pools.PopFront()
			} else {
				node = &zNode{}
			}
			node.init(int(code), p.depth+1, start, idx)
			siblings = append(siblings, node)
			// check is word
			if len(words[start]) == p.depth+1 {
				// set node is end
				node.tail = true
			}
		}

		if len(siblings) == 0 {
			continue
		}

		// step2: find usable base position
		// update nextPos
		for ; nextPos < len(dat.cells) && dat.cells[nextPos].check != 0; nextPos++ {
		}

		base := dat.findBase(siblings, nextPos)

		// step3: do insert
		dat.cells[p.index].setBase(base)
		// fmt.Printf("%d,%+v,%+v,%v\n", byte(p.code), p.index, base, dat.cells[p.index].getBase())
		for _, node := range siblings {
			index := base + node.code
			node.index = index
			// fmt.Printf("%d,%+v,%+v\n", byte(node.code), index, base)

			dat.cells[index].check = base
			if node.tail {
				node.left++
				dat.cells[index].setWord()
			}

			if !node.tail || node.right-node.left > 0 {
				nodes.PushBack(node)
			}
		}
	}

	// trim 删除多余的数据
	for i := len(dat.cells) - 1; i >= 0; i-- {
		if dat.cells[i].check != 0 || dat.cells[i].base != 0 {
			dat.cells = dat.cells[:i+1]
			break
		}
	}

	return nil
}

func (dat *DoubleArrayTrie) findBase(siblings []*zNode, nextPos int) int {
	base := nextPos - siblings[0].code
	code := siblings[len(siblings)-1].code
	for ; ; base++ {
		epos := base + code
		if epos >= len(dat.cells) {
			// resize grow
			size := 2 * epos
			t := make([]zCell, size)
			copy(t, dat.cells)
			dat.cells = t
		}

		isAll := true
		for _, node := range siblings {
			hash := base + node.code
			if dat.cells[hash].check != 0 {
				isAll = false
				break
			}
		}

		if isAll {
			return base
		}
	}
}

// Match 匹配单词,longest 为true匹配最长的一个单词,false匹配第一个
// 返回值,0没有匹配到,>0匹配到的长度
func (dat *DoubleArrayTrie) Match(str string, longest bool) int {
	cells := dat.cells
	length := len(cells)

	// 当前位置索引
	result := 0
	prevIdx := 0

	for idx := 0; idx < len(str); idx++ {
		code := str[idx]
		// for idx, code := range str {
		nextIdx := cells[prevIdx].getBase() + int(code)
		if nextIdx <= 0 || nextIdx >= length {
			return result
		}

		// 不一致,没有匹配到
		if cells[nextIdx].check != cells[prevIdx].getBase() {
			return result
		}

		// 匹配到一个单词
		if cells[nextIdx].isWord() {
			result = idx + 1
			if !longest {
				return result
			}
		}

		// 更新索引
		prevIdx = nextIdx
	}

	return result
}

func (dat *DoubleArrayTrie) dump() {
	fmt.Printf("len:%+v\n", len(dat.cells))
	fmt.Printf("%+v\t%+v\t%+v\t%v\n", "idx", "base", "check", "tail")
	for idx, cell := range dat.cells {
		fmt.Printf("%+v\t%+v\t%+v\t%v\n", idx, cell.getBase(), cell.check, cell.isWord())
	}
}

type zCell struct {
	// base用低1位标识是否是单词结束标志
	base  int
	check int
}

func (c *zCell) setBase(base int) {
	c.base |= base << 1
}

func (c *zCell) getBase() int {
	return c.base >> 1
}

func (c *zCell) setWord() {
	c.base |= 0x01
}

func (c *zCell) isWord() bool {
	return (c.base & 0x01) != 0
}

// 用于构建Trie
type zNode struct {
	prev  *zNode
	next  *zNode
	code  int
	depth int
	left  int
	right int
	index int
	tail  bool // word最后一天字符
}

func (n *zNode) init(code, depth, left, right int) {
	n.code = code
	n.depth = depth
	n.left = left
	n.right = right
	n.tail = false
	n.index = -1
}

func newList() *zList {
	return &zList{}
}

// 双向非循环队列
type zList struct {
	head *zNode
	tail *zNode
	size int
}

func (l *zList) Len() int {
	return l.size
}

func (l *zList) PushBack(node *zNode) {
	if l.tail != nil {
		l.tail.next = node
		l.tail = node
	} else {
		l.head = node
		l.tail = node
	}
	l.size++
}

func (l *zList) PopFront() *zNode {
	head := l.head
	l.head = l.head.next
	if l.head == nil {
		l.tail = nil
	}

	l.size--
	return head
}
