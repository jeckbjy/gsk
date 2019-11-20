package bi

import (
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	url := "https://use-your-test-url"
	if err := Init(&Options{URL: url, Wait: time.Second * 10}); err != nil {
		t.Fatal(err)
	}

	err := Send("login", M{"device_id": "test"})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("send")
	}

	Stop()
}

func TestReflect(t *testing.T) {
	type Login struct {
		ElapsedSecs int
		Result      int
		RetryTimes  int
		Url         string
	}

	b1 := &Login{ElapsedSecs: 1, Result: 1, RetryTimes: 0, Url: "https://test"}
	event, params, _ := Reflect(b1)
	t.Log(event, params)

	type MemberCard struct {
		Reason        string
		Step          int
		TransactionId string
	}

	b2 := &MemberCard{Reason: "popup", Step: 1, TransactionId: "test-id"}
	e2, p2, _ := Reflect(b2)
	t.Log(e2, p2)
}
