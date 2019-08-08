package backoff

import "time"

var (
	defaultFactor   float64 = 2
	defaultJitter           = false
	defaultMinDelay         = 100 * time.Millisecond
	defaultMaxDelay         = 2 * time.Second
)

// https://github.com/grpc/grpc-go/blob/master/internal/backoff/backoff.go
type Exponential struct {
	Attempts float64
	Factor   float64
	Jitter   float64
}
