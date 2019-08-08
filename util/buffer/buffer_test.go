package buffer

import (
	"encoding/json"
	"io"
	"testing"
)

func TestBuffer_Read(t *testing.T) {
	b := New()
	b.Append([]byte("hello world"))
	b.Seek(0, io.SeekStart)
	d := make([]byte, 5)
	if _, err := b.Read(d); err != nil || string(d) != "hello" {
		t.Errorf("read fail:%+v", err)
	}
}

func TestBuffer_ReadFail(t *testing.T) {
	b := New()
	b.Append([]byte("hello world"))
	b.Seek(0, io.SeekStart)
	d := make([]byte, 30)
	if _, err := b.Read(d); err == nil {
		t.Errorf("expect read overflow err")
	}
}

func TestBuffer_Write(t *testing.T) {
	b := New()
	b.Append([]byte("hello"))
	b.Seek(0, io.SeekStart)
	b.Write([]byte("world!"))
	if b.String() != "world!" {
		t.Errorf("write fail:%+v", b.String())
	}
}

func TestBuffer_Seek(t *testing.T) {
	b := New()
	b.Append([]byte("hello"))
	b.Append([]byte(" "))
	b.Append([]byte("world"))
	b.Seek(6, io.SeekStart)

	l := b.Len() - b.Pos()
	d := make([]byte, l)
	_, _ = b.Peek(d)
	if string(d) != "world" || b.Pos() != 6 {
		t.Errorf("seek start fail:%s", d)
	}

	// test seek back
	_, _ = b.Seek(-1, io.SeekCurrent)
	d1 := make([]byte, l+1)
	_, _ = b.Peek(d1)
	if string(d1) != " world" {
		t.Errorf("seek cur fail:%s", d1)
	}

	// test seek from end
	b.Seek(0, io.SeekEnd)
	b.Seek(7, io.SeekEnd)
	d2 := make([]byte, 7)
	b.Peek(d2)
	if string(d2) != "o world" {
		t.Errorf("seek end fail:%s", d2)
	}

}

func TestBuffer_Discard(t *testing.T) {
	b := New()
	b.Append([]byte("hello"))
	b.Append([]byte(" "))
	b.Append([]byte("world"))

	b.Seek(8, io.SeekStart)
	b.Discard()

	d := make([]byte, 2)
	b.Peek(d)
	if string(d) != "rl" {
		t.Errorf("discard fail:%s", d)
	}
}

func TestBuffer_Split(t *testing.T) {
	b := New()
	b.Append([]byte("hello"))
	b.Append([]byte(" "))
	b.Append([]byte("world"))

	b.Seek(8, io.SeekStart)

	s := b.Split()
	if string(s.String()) != "hello wo" || string(b.String()) != "rld" {
		t.Errorf("split fail:new=%s, old=%s", s.String(), b.String())
	}
}

func TestBuffer_Split2(t *testing.T) {
	b := New()
	b.Append([]byte("hello"))
	b.Append([]byte(" "))
	b.Append([]byte("world"))

	b.Seek(3, io.SeekStart)

	s := b.Split()
	if string(s.String()) != "hel" || string(b.String()) != "lo world" {
		t.Errorf("split fail:new=%s, old=%s", s.String(), b.String())
	}
}

func TestBuffer_ReadByte(t *testing.T) {
	b := New()
	b.Append([]byte("hello"))
	b.Append([]byte(" "))
	b.Append([]byte("world"))

	b.Seek(8, io.SeekStart)
	d, _ := b.ReadByte()
	if d != 'r' {
		t.Errorf("read byte fail:%+v", d)
	}
}

func TestBuffer_Prepend(t *testing.T) {
	b := New()
	b.Append([]byte("world"))
	b.Prepend([]byte(" "))
	b.Prepend([]byte("hello"))
	if b.String() != "hello world" {
		t.Errorf("prepend fail:%s", b.String())
	}
}

func TestBuffer_Json(t *testing.T) {
	a := make(map[string]string)
	a["aa"] = "aa"

	b := New()
	e := json.NewEncoder(b)
	if err := e.Encode(&a); err != nil {
		t.Errorf("err:%+v", err)
	}

	t.Logf("data:%+v", b.String())
}
