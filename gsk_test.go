package gsk

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	s.Run()
}
