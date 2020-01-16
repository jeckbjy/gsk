package cache

func NewLRU(capacity int) *XRU {
	c := &XRU{hook: xruPushFront}
	c.init(capacity)
	return c
}

func NewMRU(capacity int) *XRU {
	c := &XRU{hook: xruPushBack}
	c.init(capacity)
	return c
}

type xruNode struct {
	prev  *xruNode
	next  *xruNode
	key   interface{}
	value interface{}
}

type xruList struct {
	head *xruNode
	tail *xruNode
}

// LRU/MRU缓存,Get/Put都会移动节点,复杂度都是O(1)
// LRU:每次把新节点放入链表头,溢出时驱逐链表尾节点
// MRU:每次把新节点插入链表尾,溢出时驱逐链表尾节点
type XRU struct {
	base
	list xruList
	dict map[interface{}]*xruNode
	hook func(*xruList, *xruNode) // lru:pushFront,mru:pushBack
}

func (c *XRU) init(capacity int) {
	if capacity < 2 {
		panic("capacity must be greater than 1")
	}
	c.capacity = capacity
	c.dict = make(map[interface{}]*xruNode)
}

func (c *XRU) Keys() []interface{} {
	c.RLock()
	defer c.RUnlock()
	if len(c.dict) == 0 {
		return nil
	}

	keys := make([]interface{}, 0, len(c.dict))
	for key, _ := range c.dict {
		keys = append(keys, key)
	}
	return keys
}

func (c *XRU) Values() []interface{} {
	c.RLock()
	defer c.RUnlock()
	if len(c.dict) == 0 {
		return nil
	}

	values := make([]interface{}, 0, len(c.dict))
	for _, node := range c.dict {
		values = append(values, node.value)
	}
	return values
}

func (c *XRU) Has(key interface{}) bool {
	c.RLock()
	_, ok := c.dict[key]
	c.RUnlock()
	return ok
}

func (c *XRU) Put(key, value interface{}) error {
	c.Lock()
	var node *xruNode
	if n, ok := c.dict[key]; !ok {
		if len(c.dict) >= c.capacity {
			// 驱逐最后一个,限制要求capacity必须大于1,保证list不会为空
			tail := c.list.tail
			c.list.tail = tail.prev
			tail.prev = nil
			node = tail
			node.prev = nil
			node.next = nil
		}

		if node == nil {
			node = &xruNode{}
		}
		node.key = key
		c.dict[key] = node
	} else {
		node = n
	}

	node.value = value

	c.hook(&c.list, node)
	c.Unlock()
	return nil
}

func (c *XRU) Get(key interface{}) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()
	if node, ok := c.dict[key]; ok {
		c.hook(&c.list, node)
		return node.value, nil
	}

	return nil, ErrNotFound
}

func (c *XRU) Remove(key interface{}) interface{} {
	c.Lock()
	var result interface{}
	if node, ok := c.dict[key]; ok {
		result = node.value
	}
	c.Unlock()
	return result
}

func (c *XRU) Clear() {
	c.Lock()
	for node := c.list.head; node != nil; {
		temp := node
		node = node.next
		temp.prev = nil
		temp.next = nil
	}
	c.list.head = nil
	c.list.tail = nil
	c.dict = make(map[interface{}]*xruNode)
	c.Unlock()
}

func (c *XRU) Len() int {
	c.RLock()
	l := len(c.dict)
	c.RUnlock()
	return l
}

func xruPushFront(l *xruList, n *xruNode) {
	if l.head == nil {
		l.head = n
		l.tail = n
	} else {
		n.next = l.head
		l.head.prev = n
		l.head = n
	}
}

func xruPushBack(l *xruList, n *xruNode) {
	if l.tail == nil {
		l.head = n
		l.tail = n
	} else {
		n.prev = l.tail
		l.tail.next = n
		l.tail = n
	}
}
