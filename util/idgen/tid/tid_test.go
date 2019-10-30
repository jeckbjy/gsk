package tid

import (
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(id.String())
	}
}

func TestTime(t *testing.T) {
	st := time.Date(2016, time.January, 0, 0, 0, 0, 0, time.UTC)
	s1 := time.Date(2019, time.January, 0, 0, 0, 0, 0, time.UTC)
	nn := time.Now()
	tt := nn.Sub(st).Nanoseconds() / int64(time.Millisecond)
	t1 := nn.Sub(s1).Nanoseconds() / int64(time.Millisecond)
	t.Log(tt)
	t.Log(t1)
}
