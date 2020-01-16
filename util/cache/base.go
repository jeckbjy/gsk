package cache

import "sync"

type base struct {
	sync.RWMutex
	capacity int
}
