// doc: https://github.com/ionous/container

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Copyright 2017 - ionous. Modified to create intrusive linked lists.

// Package inlist implements an intrusive doubly linked list.
//
// To iterate over a list (where l is a *List):
//	for e := l.Front(); e != nil; e = e.Next() {
//		// do something with e
//	}
//
// 修改了部分接口，以便和标准的一致
// 只有MoveFrontList和MoveBackList和标准不一样，因为标准是拷贝，这里是移动合并
// 增加PopFront和PopBack接口
package inlist

// Element impl Intrusive
type Element struct {
	Hook
	Value interface{}
}

// NewElement creates an intrusive Element to store the passed value.
func NewElement(v interface{}) Intrusive {
	return &Element{Value: v}
}

// Value return value of element
func Value(i Intrusive) interface{} {
	return i.(*Element).Value
}

// Next old code,do not need!!!
func Next(e Intrusive) (ret Intrusive) {
	return e.Next()
}

// Prev old code,do not need!!!
func Prev(e Intrusive) (ret Intrusive) {
	return e.Prev()
}

// Intrusive interface of Element
type Intrusive interface {
	Prev() Intrusive
	Next() Intrusive
	List() *List
	// use for internal
	getPrev() Intrusive
	getNext() Intrusive
	setPrev(e Intrusive)
	setNext(e Intrusive)
	getList() *List
	setList(l *List)
}

type Hook struct {
	prev, next Intrusive
	list       *List
}

func (h *Hook) Prev() Intrusive {
	if p := h.prev; h.list != nil && p != &h.list.root {
		return p
	}

	return nil
}

func (h *Hook) Next() Intrusive {
	if p := h.next; h.list != nil && p != &h.list.root {
		return p
	}

	return nil
}

func (h *Hook) List() *List {
	return h.list
}

func (h *Hook) getPrev() Intrusive {
	return h.prev
}

func (h *Hook) setPrev(e Intrusive) {
	h.prev = e
}

func (h *Hook) getNext() Intrusive {
	return h.next
}

func (h *Hook) setNext(e Intrusive) {
	h.next = e
}

func (h *Hook) getList() *List {
	return h.list
}

func (h *Hook) setList(l *List) {
	h.list = l
}

type List struct {
	root Hook
	len  int
}

func (l *List) Init() *List {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

func New() *List {
	return new(List).Init()
}

func (l *List) Len() int {
	return l.len
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List) Front() Intrusive {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List) Back() Intrusive {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero List value.
func (l *List) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *List) insert(e, at Intrusive) Intrusive {
	n := at.getNext()
	at.setNext(e)
	e.setPrev(at)
	e.setNext(n)
	n.setPrev(e)
	e.setList(l)
	l.len++
	return e
}

// remove removes e from its list, decrements l.len, and returns e.
func (l *List) remove(e Intrusive) Intrusive {
	e.getPrev().setNext(e.getNext())
	e.getNext().setPrev(e.getPrev())
	e.setNext(nil)
	e.setPrev(nil)
	e.setList(nil)
	l.len--
	return e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *List) Remove(e Intrusive) bool {
	if e.getList() == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
		return true
	}

	return false
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List) PushFront(e Intrusive) Intrusive {
	l.lazyInit()
	return l.insert(e, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List) PushBack(e Intrusive) Intrusive {
	l.lazyInit()
	return l.insert(e, l.root.prev)
}

// PopFront remove and return front element
func (l *List) PopFront() Intrusive {
	e := l.Front()
	l.remove(e)
	return e
}

// PopBack remove and return back element
func (l *List) PopBack() Intrusive {
	e := l.Back()
	l.remove(e)
	return e
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List) InsertBefore(e Intrusive, mark Intrusive) Intrusive {
	if mark.getList() != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insert(e, mark.getPrev())
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List) InsertAfter(e Intrusive, mark Intrusive) Intrusive {
	if mark.getList() != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insert(e, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List) MoveToFront(e Intrusive) {
	if e.getList() != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.insert(l.remove(e), &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List) MoveToBack(e Intrusive) {
	if e.getList() != l || l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.insert(l.remove(e), l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List) MoveBefore(e, mark Intrusive) {
	if e.getList() != l || e == mark || mark.getList() != l {
		return
	}
	l.insert(l.remove(e), mark.getPrev())
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List) MoveAfter(e, mark Intrusive) {
	if e.getList() != l || e == mark || mark.getList() != l {
		return
	}
	l.insert(l.remove(e), mark)
}

// MoveBackList moves all elements from other to the end of this list.
// diff with MoveBackList, because that is copy, here is move
func (l *List) MoveBackList(other *List) {
	if l != other {
		l.lazyInit()
		for e := other.Front(); e != nil; {
			n := e.Next()
			l.insert(e, l.root.prev)
			e = n
		}
		other.Init()
	}
}

// MoveFrontList moves all elements from other to the front of this list.
// diff with PushFrontList, because that is copy, here is move
func (l *List) MoveFrontList(other *List) {
	if l != other {
		l.lazyInit()
		for e := other.Back(); e != nil; {
			p := e.Prev()
			l.insert(e, &l.root)
			e = p
		}
		other.Init()
	}
}
