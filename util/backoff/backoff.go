package backoff

import "time"

const (
	Stop = -1
)

// BackOff 接口定义
type BackOff interface {
	// return -1 mean stop
	Next() time.Duration
	Reset()
}
