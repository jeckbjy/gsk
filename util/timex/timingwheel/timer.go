package timingwheel

// 基于回调的定时器
type Timer struct {
	engine  *TimingWheel //
	list    *bucket      // 非空则表示还是等待计算超时
	prev    *Timer       //
	next    *Timer       //
	expired int64        // 过期时间
	period  int64        // ticker使用
	task    func()       // 回调任务
}

func (t *Timer) Stop() bool {
	if t.engine != nil {
		return t.engine.Remove(t)
	}

	return false
}
