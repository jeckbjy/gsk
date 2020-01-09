package timingwheel

const (
	slotPow = 10           // 2^slotPow,默认为10
	slotMax = 1 << slotPow // 每个wheel中slot最大个数
)

func newWheel(index int) *wheel {
	w := &wheel{}
	w.init(index)
	return w
}

type wheel struct {
	slots   []*bucket // 所有桶
	index   int       // 当前索引
	offset  int       // 偏移位数
	maximum int64     // 最大取值
}

func (w *wheel) init(index int) {
	w.index = 0
	w.offset = index * slotPow
	w.maximum = int64(uint64(1) << ((index + 1) * slotPow))
	for i := 0; i < slotMax; i++ {
		w.slots = append(w.slots, newBucket())
	}
}

// 返回当前指向的桶
func (w *wheel) Current() *bucket {
	return w.slots[w.index]
}

// 向前前进一格,到最后一格则归零,并返回true
func (w *wheel) Step() bool {
	w.index++
	if w.index >= slotMax {
		w.index = 0
		return true
	}

	return false
}

// 添加到正确的桶当中
func (w *wheel) Push(t *Timer, delta int64) {
	off := int(delta>>w.offset) - 1
	index := (off + w.index) % slotMax
	w.slots[index].Push(t)
}

//func (self *Wheel) Push(timer *Timer, delta uint64) {
//	off := (delta >> self.timeOff) - 1
//	index := (int(off) + self.index) % len(self.slots)
//	self.slots[index].PushBack(timer)
//}

//import "github.com/jeckbjy/fairy/container/inlist"

//const (
//	TIME_INTERVAL = 1             // 默认时间间隔
//	WHEEL_NUM     = 3             // 初始wheel个数(2^30)，可以扩展
//	SLOT_POW      = 10            // 2^SLOT_POW,默认10
//	SLOT_MAX      = 1 << SLOT_POW // 个数
//)
//
//type Wheel struct {
//	slots   []*inlist.List // 桶
//	index   int            // 当前slot循环索引
//	timeOff uint           // shift offset
//	timeMax uint64         // 区间最大值
//}
//
//func (self *Wheel) Create(index int) {
//	self.index = 0
//	self.timeOff = uint(index * SLOT_POW)
//	self.timeMax = uint64(1) << uint((index+1)*SLOT_POW)
//	for i := 0; i < SLOT_MAX; i++ {
//		self.slots = append(self.slots, inlist.New())
//	}
//}
//
//func (self *Wheel) Current() *inlist.List {
//	return self.slots[self.index]
//}
//
//func (self *Wheel) Step() bool {
//	self.index++
//	if self.index >= len(self.slots) {
//		self.index = 0
//		return true
//	}
//
//	return false
//}
//
//func (self *Wheel) Push(timer *Timer, delta uint64) {
//	off := (delta >> self.timeOff) - 1
//	index := (int(off) + self.index) % len(self.slots)
//	self.slots[index].PushBack(timer)
//}
