package zset

import (
	"fmt"
	"testing"
)

func toStr(x interface{}) string {
	return fmt.Sprintf("%+v", x)
}

func create(max int) *SortedSet {
	zs := New()
	for i := int64(1); i <= int64(max); i++ {
		zs.Insert(toStr(i), i)
	}

	return zs
}

func TestNew(t *testing.T) {
	zs := create(10)

	for i := uint64(1); i < 10; i++ {
		if zs.GetRank(toStr(i)) != i {
			t.Errorf("GetRank fail, key=%v", i)
		}
	}

	zs.Delete("5")

	if zs.Len() != 9 {
		t.Errorf("zset len bad")
	}

	for i := uint64(1); i < 4; i++ {
		if zs.GetRank(toStr(i)) != i {
			t.Errorf("GetRank fail, key=%v", i)
		}
	}

	for i := uint64(6); i < 10; i++ {
		if zs.GetRank(toStr(i)) != (i - 1) {
			t.Errorf("GetRank fail, key=%v", i)
		}
	}

	zs.Scan(func(rank uint64, el *Element) {
		t.Logf("scan rank=%v, key=%v", rank, el.Key)
	})
}

func TestDelete(t *testing.T) {
	zs := create(1)
	zs.Delete("1")
	if zs.Len() != 0 {
		t.Errorf("bad count")
	}
}

func TestMax(t *testing.T) {
	zs := create(10)
	zs.SetMax(5)

	if zs.Len() != 5 {
		t.Errorf("zset max fail")
	}
}

func TestDesending(t *testing.T) {
	zs := create(10)
	zs.SetDescending()
	zs.SetMax(5)

	zs.Scan(func(rank uint64, el *Element) {
		t.Logf("zset rank=%v, key=%v, score=%v", rank, el.Key, el.Score)
	})
}

func TestLargeRange(t *testing.T) {
	max := 1000000
	zs := create(max)

	for i := 1; i < 10; i++ {
		// t.Logf("key=%+v, rank=%+v", i, zs.GetRank(util.ConvStr(i)))
		if zs.GetRank(toStr(i)) != uint64(i) {
			t.Errorf("GetRank fail, key=%v", i)
		}
	}
}

func TestLoadSave(t *testing.T) {
	zs := create(20)
	path := "./dump.db"
	zs.SaveFile(path)

	// load
	zs1 := New()
	zs1.LoadFile(path)

	if zs1.Len() != 10 {
		t.Errorf("load fail:len=%+v", zs1.Len())
	}

	zs1.Scan(func(rank uint64, e *Element) {
		t.Logf("rank=%+v,key=%+v", rank, e.Key)
	})
}
