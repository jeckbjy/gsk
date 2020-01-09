package arrmap

import (
	"sort"
)

type strItem struct {
	key string
	val interface{}
}

// 递增有序,相当于map[string]interface{},使用数组保存
// https://developer.android.com/reference/android/support/v4/util/ArrayMap
type StringMap struct {
	items []strItem
}

func (m *StringMap) IsEmpty() bool {
	return len(m.items) == 0
}

func (m *StringMap) Len() int {
	return len(m.items)
}

func (m *StringMap) KeyAt(index int) string {
	return m.items[index].key
}

func (m *StringMap) ValueAt(index int) interface{} {
	return m.items[index].val
}

func (m *StringMap) IndexOf(key string) int {
	size := len(m.items)
	if size == 0 {
		return -1
	}

	index := sort.Search(len(m.items), func(i int) bool { return key < m.items[i].key })
	if index == size || m.items[index].key != key {
		return -1
	}

	return index
}

func (m *StringMap) ContainsKey(key string) bool {
	return m.IndexOf(key) != -1
}

func (m *StringMap) ContainsValue(value interface{}) bool {
	for _, item := range m.items {
		if item.val == value {
			return true
		}
	}

	return false
}

func (m *StringMap) Get(key string) (interface{}, bool) {
	index := m.IndexOf(key)
	if index == -1 {
		return nil, false
	}

	return m.items[index].val, true
}

func (m *StringMap) Set(key string, value interface{}) {
	item := strItem{key: key, val: value}
	size := len(m.items)
	if size == 0 {
		m.items = append(m.items, item)
		return
	}

	idx := sort.Search(size, func(i int) bool { return key < m.items[i].key })
	m.items = append(m.items, item)
	if idx < size {
		copy(m.items[idx+1:], m.items[idx:])
		m.items[idx] = item
	}
}

func (m *StringMap) Remove(key string) interface{} {
	index := m.IndexOf(key)
	if index == -1 {
		return nil
	}
	value := m.items[index].val
	m.items = append(m.items[:index], m.items[index+1:]...)
	return value
}

func (m *StringMap) RemoveAt(index int) interface{} {
	value := m.items[index].val
	m.items = append(m.items[:index], m.items[index+1:]...)
	return value
}

func (m *StringMap) SetValueAt(index int, value interface{}) {
	m.items[index].val = value
}

func (m *StringMap) Clear() {
	m.items = nil
}
