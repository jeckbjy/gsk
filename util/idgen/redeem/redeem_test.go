package redeem

import (
	"math/rand"
	"testing"
	"time"
)

func TestInject(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := uint64(10000); i < 10100; i++ {
		x := injectBits(i, 4)
		y := extractBits(x, 4)
		if i != y {
			t.Fatal("not equal", i, x, y)
		} else {
			t.Log(i, x)
		}
	}
}

func TestEncode(t *testing.T) {
	g := NewEx(nil, 2, 5)
	tabId := uint64(100)

	a := g.Encode(tabId, 100)
	t.Log(a, g.Generate(tabId, 100))

	start := uint64(1)
	for i := 0; i < 5; i++ {
		t.Logf("----------%+v----------", start)
		for s := start; s < start+10; s++ {
			a := g.Generate(tabId, uint64(s))
			b, err := g.Verify(a)
			if err != nil {
				t.Fatal(err, s)
			} else if b != tabId {
				t.Fatal("not equal", s, b)
			} else {
				t.Log(s, "->", a)
			}
		}

		start *= 100
	}
}

func TestABC(t *testing.T) {
	abc := NewAlphabet("ABCDEFGHJKLMNPQRSTUVWXYZ", 1)
	x := uint64(1111111)
	a := abc.Encode(x)
	b, err := abc.Decode(a)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("ok", a, b)
	}
}
