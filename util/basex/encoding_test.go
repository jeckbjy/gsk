package basex

import (
	"testing"
	"time"
)

func TestEncoding(t *testing.T) {
	b64, _ := NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	n := uint64(time.Now().UnixNano())
	s := Encode(n)
	x := Decode(s)
	if x != n {
		t.Error("not equal", n, s, x)
	} else {
		t.Log(s)
	}

	s1 := b64.Encode(n)
	x1 := b64.Decode(s1)
	if x1 != n {
		t.Error("not equal", s1, x1)
	} else {
		t.Log(s1)
	}

	for i := uint64(0); i < 1000; i++ {
		s := Encode(i)
		x := Decode(s)
		//t.Log(s)
		if i != x {
			t.Error("not equal", i)
		}
	}

	t.Log("ok")
}

func TestShuffle(t *testing.T) {
	abc := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	seed := time.Now().Unix()
	x := Shuffle(abc, uint64(seed))
	t.Log(x)
	y := Shuffle(abc, uint64(seed))
	t.Log(x)

	if len(x) != len(abc) || len(y) != len(abc) {
		t.Error("bad length")
	} else {
		t.Log("length ok")
	}

	if x != y {
		t.Error("not same")
	} else {
		t.Log("ok")
	}
}

func TestCheckUnique(t *testing.T) {
	abc := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	if CheckUnique(abc) {
		t.Log("is unique")
	} else {
		t.Error("not unique")
	}

	def := "1233"
	if !CheckUnique(def) {
		t.Log("not unique")
	} else {
		t.Error("bad")
	}
}
