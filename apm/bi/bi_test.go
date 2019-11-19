package bi

import "testing"

func TestSend(t *testing.T) {
	if err := Init(&Options{URL: ""}); err != nil {
		t.Fatal(err)
	}

	err := Send("", M{"key": "aa"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReflect(t *testing.T) {
	type FooFFxx struct {
		Name  string `bi:"name"`
		Value int
	}

	tt := &FooFFxx{Name: "test", Value: 1}

	event, params, err := Reflect(tt)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(event, params)
	}

	t2 := FooFFxx{Name: "tt", Value: 2}
	e2, p2, _ := Reflect(t2)
	t.Log(e2, p2)
}
