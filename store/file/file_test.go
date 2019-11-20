package file

import (
	"testing"

	"github.com/jeckbjy/gsk/store"
)

func TestFileStore(t *testing.T) {
	s := New()
	key := "config"
	val := "test"
	err := s.Put(nil, key, []byte(val))
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := s.Exists(nil, key); err != nil || !ok {
		t.Fatal("not exists")
	} else {
		t.Log("put ok")
	}

	kv, err := s.Get(nil, key)
	if err != nil && err != store.ErrNotFound {
		t.Fatal(err)
	}

	if kv == nil {
		t.Fatal("no data")
	}

	if string(kv.Value) != val {
		t.Fatal("not equal")
	}

	t.Logf("put value, %s", kv.Value)

	if err := s.Delete(nil, key); err != nil {
		t.Fatal(err)
	}

	if ok, _ := s.Exists(nil, key); ok {
		t.Fatal("delete fail")
	} else {
		t.Log("delete ok")
	}

	t.Log("list all-------")
	files, err := s.List(nil, "", store.KeyOnly())
	if err != nil && err != store.ErrNotFound {
		t.Fatal(err)
	}

	for _, f := range files {
		t.Log(f.Key, string(f.Value))
	}
}
