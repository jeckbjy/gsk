package breaker

import "time"

var gTimeClock = newTimeClock()

func newTimeClock() _Clock {
	return &timeClock{}
}

func newMockClock(delta time.Duration, start time.Time, autoAdd bool) *mockClock {
	return &mockClock{delta: delta, now: start, autoAdd: autoAdd}
}

type _Clock interface {
	Now() time.Time
}

type timeClock struct {
}

func (c *timeClock) Now() time.Time {
	return time.Now()
}

type mockClock struct {
	now     time.Time
	delta   time.Duration
	autoAdd bool
}

func (m *mockClock) Add(delta time.Duration) {
	m.now = m.now.Add(delta)
}

func (m *mockClock) Now() time.Time {
	if m.autoAdd {
		m.now = m.now.Add(m.delta)
	}
	return m.now
}
