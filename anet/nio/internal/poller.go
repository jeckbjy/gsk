package internal

const maxEventNum = 1024

type Operation int

const (
	OP_READ  Operation = 0x04
	OP_WRITE           = 0x08
	OP_RW              = OP_READ | OP_WRITE
)

type poller interface {
	Open() error
	Close() error

	Wakeup() error
	Wait(s *Selector, cb SelectCB, msec int) error

	Add(fd uintptr, ops Operation) error
	Del(fd uintptr, ops Operation) error
	Mod(fd uintptr, old, ops Operation) error
}
