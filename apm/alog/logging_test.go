package alog

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetFrame(t *testing.T) {
	f := getFrame(0)
	t.Log(f.File, filepath.Base(f.File), f.Line, filepath.Base(f.Function), f.Function)
}

func TestPrint(t *testing.T) {
	ss := fmt.Sprintf("%-10s aa", "aa")
	t.Log(ss)
}

func TestExecName(t *testing.T) {
	t.Log(filepath.Base(os.Args[0]))
}

func TestDateFormat(t *testing.T) {
	df := DateFormat{}
	df.Parse("yyyy-MM-ddTHH:mm:ssz")
	str := df.Format(time.Now())
	t.Log(str)
}

func TestLayout(t *testing.T) {
	l := &Layout{}
	if err := l.Parse("[%D %6p %-25L] %t"); err != nil {
		t.Error(err)
	}
	msg := &Entry{
		Frame: getFrame(0),
		Time:  time.Now(),
		Level: LevelDebug,
		Text:  "test",
	}

	s := l.Format(msg)
	t.Log(s)
}

func TestLogging(t *testing.T) {
	a := 1
	b := "hello"
	Trace(a, b)
	time.Sleep(time.Second)
}
