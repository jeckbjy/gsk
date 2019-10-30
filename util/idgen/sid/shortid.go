package sid

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	randc "crypto/rand"
	randm "math/rand"
)

var ErrTimeOverflow = errors.New("shortid time overflow")
var maxTimestamp = uint64(1) << 40

var gDefault = MustNew(0, DefaultABC, 1)

func GetDefault() *Shortid {
	return (*Shortid)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&gDefault))))
}

func SetDefault(sid *Shortid) {
	target := (*unsafe.Pointer)(unsafe.Pointer(&gDefault))
	source := unsafe.Pointer(sid)
	atomic.SwapPointer(target, source)
}

func Generate() (string, error) {
	return gDefault.Generate()
}

func MustGenerate() string {
	id, err := Generate()
	if err != nil {
		panic(err)
	}

	return id
}

func New(worker uint, alphabet string, seed uint64) (*Shortid, error) {
	abc, err := NewAbc(alphabet, seed)
	if err != nil {
		return nil, err
	}

	sid := &Shortid{
		abc:     abc,
		worker:  uint64(worker),
		epoch:   time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
		mux:     &sync.Mutex{},
		ms:      0,
		counter: 0,
	}

	return sid, nil
}

func MustNew(worker uint, alphabet string, seed uint64) *Shortid {
	sid, err := New(worker, alphabet, seed)
	if err != nil {
		panic(err)
	}

	return sid
}

type Shortid struct {
	abc     Abc
	epoch   time.Time
	mux     sync.Locker
	worker  uint64
	ms      uint64
	counter uint64
}

func (s *Shortid) Generate() (string, error) {
	return s.generate(time.Now(), &s.epoch)
}

func (s *Shortid) generate(now time.Time, epoch *time.Time) (string, error) {
	// getTimeAndCounter
	s.mux.Lock()
	ms := uint64(now.Sub(*epoch).Nanoseconds() / int64(time.Millisecond))
	if ms == s.ms {
		s.counter++
	} else {
		s.counter = 0
		s.ms = ms
	}

	counter := s.counter
	s.mux.Unlock()
	if ms > maxTimestamp {
		return "", ErrTimeOverflow
	}

	// encode, time(40bits) + worker + counter
	// 高1位作为随机位
	buff := make([]rune, 0, 12)
	if s.abc.Len() == 32 {
		// 高1位随机,低4位真实数据
		buff = s.encode(buff, ms, 10, 4)
		buff = s.encode(buff, s.worker, 0, 4)
		if counter > 0 {
			buff = s.encode(buff, counter, 0, 4)
		}
	} else {
		// 高1位随机,低5位真实数据
		buff = s.encode(buff, ms, 8, 5)
		buff = s.encode(buff, s.worker, 0, 5)
		if counter > 0 {
			buff = s.encode(buff, counter, 0, 5)
		}
	}

	return string(buff), nil
}

func (s *Shortid) encode(buff []rune, value uint64, symbols uint, digits uint) []rune {
	if symbols == 0 {
		if value > 0 {
			symbols = uint(math.Log2(float64(value)))/digits + 1
		} else {
			symbols = 1
		}
	}

	maskH := byte(1 << digits)
	maskL := maskH - 1

	randoms := randomBytes(symbols)
	for i := uint(0); i < symbols; i++ {
		idx := (randoms[i] & maskH) | (byte(value) & maskL)
		buff = append(buff, s.abc.alphabet[idx])
		value >>= digits
	}

	return buff
}

func randomBytes(size uint) []byte {
	bytes := make([]byte, size)
	if _, err := randc.Read(bytes); err != nil {
		return bytes
	} else {
		for i := uint(0); i < size; i++ {
			bytes[i] = byte(randm.Intn(0xff))
		}
	}

	return bytes
}
