package router

import (
	"testing"

	"github.com/jeckbjy/gsk/arpc"
)

func TestContext_Next(t *testing.T) {
	m1 := func(ctx arpc.Context) error {
		t.Log("before m1")
		if err := ctx.Next(); err != nil {
			return err
		}
		t.Log("after m1")
		return nil
	}

	m2 := func(ctx arpc.Context) error {
		t.Log("before m2")
		if err := ctx.Next(); err != nil {
			return err
		}
		t.Log("after m2")
		return nil
	}

	hh := func(ctx arpc.Context) error {
		t.Log("handler")
		return nil
	}

	c := NewContext()
	c.Init(nil, nil)
	c.SetHandler(hh)
	c.SetMiddleware([]arpc.HandlerFunc{m1, m2})
	_ = c.Next()
}
