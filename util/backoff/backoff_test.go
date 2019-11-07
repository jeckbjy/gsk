package backoff

import (
	"testing"
	"time"
)

func Test1(t *testing.T) {
	b := NewExponential(
		WithMin(100*time.Millisecond),
		WithMax(10*time.Second),
		WithFactor(2),
	)

	equals(t, b.Next(), 100*time.Millisecond)
	equals(t, b.Next(), 200*time.Millisecond)
	equals(t, b.Next(), 400*time.Millisecond)
	for index := 0; index < 100; index++ {
		b.Next()
	}

	// is max
	equals(t, b.Next(), 10*time.Second)
	b.Reset()
	equals(t, b.Next(), 100*time.Millisecond)
}

func Test2(t *testing.T) {
	b := NewExponential(
		WithMin(100*time.Millisecond),
		WithMax(10*time.Second),
		WithFactor(1.5),
	)

	equals(t, b.Next(), 100*time.Millisecond)
	equals(t, b.Next(), 150*time.Millisecond)
	equals(t, b.Next(), 225*time.Millisecond)
	b.Reset()
	equals(t, b.Next(), 100*time.Millisecond)
}

func TestJitter(t *testing.T) {
	b := NewExponential(
		WithMin(100*time.Millisecond),
		WithMax(10*time.Second),
		WithFactor(2),
		WithJitter(true),
	)

	equals(t, b.Next(), 100*time.Millisecond)
	between(t, b.Next(), 100*time.Millisecond, 200*time.Millisecond)
	between(t, b.Next(), 100*time.Millisecond, 400*time.Millisecond)
	b.Reset()
	equals(t, b.Next(), 100*time.Millisecond)
}

func between(t *testing.T, actual, low, high time.Duration) {
	if actual < low {
		t.Fatalf("Got %s, Expecting >= %s", actual, low)
	}
	if actual > high {
		t.Fatalf("Got %s, Expecting <= %s", actual, high)
	}
}

func equals(t *testing.T, d1, d2 time.Duration) {
	if d1 != d2 {
		t.Fatalf("Got %s, Expecting %s", d1, d2)
	}
}
