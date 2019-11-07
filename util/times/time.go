package times

import "time"

// Now 返回毫秒值
func Now() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// NowUnix 返回秒
func NowUnix() int64 {
	return time.Now().Unix()
}
