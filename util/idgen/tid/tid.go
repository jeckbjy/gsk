package tid

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	ErrSeqIDOverflow    = errors.New("sequence id overflow")
	ErrTimeBack         = errors.New("id time back")
	ErrWorkerIDOverflow = errors.New("worker id overflow")
)

type Mode int

const (
	ModeSec  Mode = 0 // 最大峰值,秒为单位
	ModeMsec      = 1 // 最小粒度,毫秒为单位
)

const (
	maxSecondID      = 2 ^ 21 // 每秒最多1048576个
	maxMillisecondID = 2 ^ 11 // 每毫秒最多1024个
	maxWorkerID      = 2 ^ 10 //
)

var (
	defaultStartTime = time.Date(2019, time.January, 0, 0, 0, 0, 0, time.UTC)
)

var gIDGenerator *IDGen

func init() {
	gIDGenerator, _ = NewGenerator(0, false, nil)
}

func NewGenerator(nodeID uint, useSec bool, mux sync.Locker) (*IDGen, error) {
	g := &IDGen{}
	if err := g.Init(nodeID, useSec, mux); err != nil {
		return nil, err
	}

	return g, nil
}

func SetDefault(idgen *IDGen) {
	target := (*unsafe.Pointer)(unsafe.Pointer(&gIDGenerator))
	source := unsafe.Pointer(idgen)
	atomic.SwapPointer(target, source)
}

func Generate() (ID, error) {
	return gIDGenerator.Generate()
}

func MustGenerate() ID {
	id, err := gIDGenerator.Generate()
	if err != nil {
		panic(err)
	}

	return id
}

// ID生成器,time based distributed unique id
// 源自vesta-id-generator,但有有所不同,保留了秒,毫秒的设置,但是去除了版本,生成方式字段
// 又区别于snowflake,使用1位标识了时间模式(秒和毫秒)
// 需要注意的一个问题:由于强依赖时钟,时钟回调会导致ID重复,在毫秒模式下会更严重,而秒模式下因为回调时间比较短,可能并不明显
// ID生成器需要关闭NTP同步
// 当发现时钟回调时,算法应该报错
// workId:共占用10bit,可以自行定义含义,比如5位DataCenterID,5位表示机器ID
// 默认编码格式:
//				类型 workId	序列号	时间戳	符号位	占用位数
// 最大峰值(s):	0	1-10	11-31	32-62 	63		1-10-21-31
// 最小粒度(ms):	0	1-10	11-21	22-62	63		1-10-11-41
//
// https://www.callicoder.com/distributed-unique-id-sequence-number-generator/
type IDGen struct {
	mux         sync.Locker // 锁
	timestamp   uint64      // 时间戳
	sequenceID  uint64      // 自增序列
	templateID  uint64      // 模板ID
	precision   int64       // 精度
	offsetTS    uint8       // 时间戳偏移位
	sequenceMax uint64      // 最大序列值
}

func (g *IDGen) Init(nodeID uint, useSec bool, mux sync.Locker) error {
	if nodeID > maxWorkerID {
		return ErrWorkerIDOverflow
	}

	if mux == nil {
		g.mux = &sync.Mutex{}
	} else {
		g.mux = mux
	}

	if useSec {
		// 秒模式
		g.precision = int64(time.Second)
		g.sequenceMax = maxSecondID
		g.offsetTS = 32
		g.templateID = uint64(nodeID << 1)
	} else {
		// 毫秒模式
		g.precision = int64(time.Millisecond)
		g.sequenceMax = maxMillisecondID
		g.offsetTS = 22
		g.templateID = uint64(nodeID<<1) + 1
	}

	return nil
}

func (g *IDGen) Generate() (ID, error) {
	g.mux.Lock()
	defer g.mux.Unlock()

	ts := uint64(time.Now().Sub(defaultStartTime).Nanoseconds() / g.precision)
	switch {
	case ts < g.timestamp:
		// 时间发生了回退
		return 0, ErrTimeBack
	case ts > g.timestamp:
		g.timestamp = ts
		g.sequenceID = 0
	default:
		g.sequenceID++
	}

	if g.sequenceID >= g.sequenceMax {
		return 0, ErrSeqIDOverflow
	}

	return ID(ts<<g.offsetTS | g.sequenceID<<11 | g.templateID), nil
}

type ID uint64

func (id ID) Mode() Mode {
	return Mode(id & 0x01)
}

func (id ID) Time() time.Time {
	if id.Mode() == ModeSec {
		sec := int64(id>>32) + defaultStartTime.Unix()
		return time.Unix(sec, 0)
	} else {
		ms := int64(id>>22) + defaultStartTime.UnixNano()/int64(time.Millisecond)
		return time.Unix(ms/1000, (ms%1000)*1000000)
	}
}

func (id ID) SequenceID() int {
	if id.Mode() == ModeSec {
		// 21
		return int((id >> 11) & 0x1FFFFF)
	} else {
		// 11
		return int((id >> 11) & 0x7FF)
	}
}

func (id ID) NodeID() int {
	// 10
	return int((id >> 1) & 0x3FF)
}

func (id ID) IsNil() bool {
	return id == 0
}

// 以数字形式展示,年月日-时分秒-毫秒-序列号-机器ID-类型
func (id ID) String() string {
	ts := id.Time()
	seq := id.SequenceID()
	nodeID := id.NodeID()
	mode := id.Mode()
	if mode == ModeSec {
		return fmt.Sprintf("%04d%02d%02d-%02d%02d%02d-%d-%d-%d",
			ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(),
			seq, nodeID, mode)
	} else {
		return fmt.Sprintf("%04d%02d%02d-%02d%02d%02d-%04d-%d-%d-%d",
			ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Nanosecond()/1000000,
			seq, nodeID, mode)
	}
}
