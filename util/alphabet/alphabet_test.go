package alphabet

import (
	"testing"
	"time"
)

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
