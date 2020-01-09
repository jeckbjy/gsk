package timingwheel

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var gTimingWheel = New()

// 返回全局默认的
func Get() *TimingWheel {
	return gTimingWheel
}

// 创建定时器
func NewTimer(expired int64, task func()) *Timer {
	return gTimingWheel.NewTimer(expired, task)
}

// 新建TimingWheel,可指定Interval,默认为1msec
func New(opts ...Option) *TimingWheel {
	o := Options{interval: 1}
	for _, fn := range opts {
		fn(&o)
	}

	tw := &TimingWheel{interval: o.interval}
	tw.start()
	return tw
}

// Hierarchical Time Wheel
// https://www.cnblogs.com/zhongwencool/p/timing_wheel.html
// TimingWheel 拥有两个协程,第一个用于计算timer过期,第二个用于执行timer
type TimingWheel struct {
	mux       sync.Mutex //
	wheels    []*wheel   // 时间轮
	interval  int64      // 时间间隔,精度毫秒
	timestamp int64      // 当前时间戳,单位毫秒
	maximum   int64      // 当前能保存的最大范围
	count     int        // 定时器个数
	exit      int32      // 是否需要退出
	exec      executor   // 单独线程执行
}

func (tw *TimingWheel) SetInterval(v int64) {
	tw.interval = v
}

func (tw *TimingWheel) NewTimer(expired int64, task func()) *Timer {
	t := &Timer{expired: expired, task: task}
	tw.Add(t)
	return t
}

// 添加定时器
func (tw *TimingWheel) Add(timer *Timer) {
	tw.mux.Lock()
	if timer.expired > tw.timestamp {
		timer.engine = tw
		tw.push(timer)
		tw.count++
	}
	tw.mux.Unlock()
}

// 删除定时器
func (tw *TimingWheel) Remove(timer *Timer) bool {
	tw.mux.Lock()
	result := false
	if timer.list != nil {
		result = true
		timer.list.Remove(timer)
		tw.count--
	}
	tw.mux.Unlock()

	return result
}

func (tw *TimingWheel) newWheel() {
	wheel := newWheel(len(tw.wheels))
	tw.maximum = wheel.maximum
	tw.wheels = append(tw.wheels, wheel)
}

func (tw *TimingWheel) start() {
	tw.exit = 0
	tw.interval = 1
	tw.timestamp = time.Now().UnixNano() / int64(time.Millisecond)

	// 默认只初始化3个wheel,可用范围2^30 * tw.tick毫秒,共占用3*1024个slot
	for i := 0; i < 3; i++ {
		tw.newWheel()
	}

	tw.exec.Start()
	go tw.Run()
}

func (tw *TimingWheel) Stop() {
	atomic.StoreInt32(&tw.exit, 1)
	tw.exec.Stop()
}

//
func (tw *TimingWheel) Run() {
	for atomic.LoadInt32(&tw.exit) == 0 {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		cur := tw.Tick(now)
		sleep := now - cur
		if sleep == 0 {
			sleep = tw.interval
		}
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}
}

// Tick 根据当前时间戳计算过期的timer
// 原则上每次调用now都应该递增,如果发生时间回调,将会rebuild所有timer,仅在调试环境下才能发生,
// rebuild是一个相对耗时操作,会遍历所有桶
func (tw *TimingWheel) Tick(now int64) int64 {
	pendings := bucket{}

	tw.mux.Lock()
	if tw.count == 0 {
		tw.timestamp = now
	} else {
		tw.doTick(&pendings, now)
	}

	timestamp := tw.timestamp
	tw.mux.Unlock()

	tw.exec.Post(&pendings)
	return timestamp
}

func (tw *TimingWheel) doTick(pendings *bucket, now int64) {
	if tw.timestamp <= now {
		ticks := (now - tw.timestamp) / tw.interval
		if ticks > 0 {
			tw.timestamp += tw.interval * ticks
			for i := int64(0); i < ticks; i++ {
				pendings.merge(tw.wheels[0].Current())
				tw.cascade(pendings)
			}
		}
	} else {
		tw.timestamp = now
		// 发生时间回调,重新构建所有时间,需要重新计算所有timer
		for _, wheel := range tw.wheels {
			wheel.index = 0
			for _, slot := range wheel.slots {
				pendings.merge(slot)
			}
		}

		// 重新计算时间
		for iter := pendings.Front(); iter != nil; {
			timer := iter
			iter = iter.next
			if timer.expired >= now {
				pendings.Remove(timer)
				tw.push(timer)
			}
		}
	}
	pendings.unlink()
	tw.count -= pendings.Len()
}

// cascade 前进1个tick,并返回过期的timer,存于pendings中
func (tw *TimingWheel) cascade(pendings *bucket) {
	for i := 0; i < len(tw.wheels); i++ {
		if !tw.wheels[i].Step() {
			break
		}

		if i+1 == len(tw.wheels) {
			// 溢出,创建新的wheel
			tw.newWheel()
			break
		}

		// rehash next wheel
		slots := tw.wheels[i+1].Current()
		for iter := slots.Front(); iter != nil; {
			timer := iter
			iter = timer.next
			slots.Remove(timer)

			if timer.expired <= tw.timestamp {
				pendings.Push(timer)
			} else {
				tw.push(timer)
			}
		}
	}
}

// 计算并添加到正确的wheel中
// 外部调用需要保证正确expired一定大于timestamp
func (tw *TimingWheel) push(timer *Timer) {
	var delta int64
	if tw.interval != 1 {
		delta = int64(math.Ceil(float64((timer.expired - tw.timestamp) / tw.interval)))
	} else {
		delta = timer.expired - tw.timestamp
	}

	// 溢出则动态添加wheel
	for delta > tw.maximum {
		tw.newWheel()
	}

	for i := 0; i < len(tw.wheels); i++ {
		wheel := tw.wheels[i]
		if delta < wheel.maximum {
			wheel.Push(timer, delta)
			break
		}
	}
}
