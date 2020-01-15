package errorx

import (
	"errors"
	"net/http"
	"testing"
)

func TestCaller(t *testing.T) {
	s := getCaller(1)
	t.Log(s)
}

func TestNew(t *testing.T) {
	err := New(errors.New("test"), http.StatusInternalServerError, "index=%+v", 123)
	t.Log(err)
	t.Log(err.Debug())
	err1 := Unauthorized("token=%+v", 123)
	t.Log(err1.Debug())
}
