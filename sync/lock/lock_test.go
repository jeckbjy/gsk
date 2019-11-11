package lock

import (
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	go Foo(t, 1)
	go Foo(t, 2)
	time.Sleep(time.Second * 2)
}

func Foo(t *testing.T, index int) {
	l, err := Lock("test_key", Wait(time.Millisecond*500, 1))
	// not acquire locker
	if err != nil {
		t.Log(err, index)
		return
	}
	t.Log("Acquire lock", index)
	time.Sleep(time.Second)
	l.Unlock()
	t.Log("free lock")
}
