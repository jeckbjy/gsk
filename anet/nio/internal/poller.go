package internal

import "syscall"

const maxEventNum = 1024

const (
	EventRead  = evRead
	EventWrite = evWrite
	EventError = evError
)

type Event struct {
	Fd     fd_t
	Events int
}

func (e *Event) Readable() bool {
	return e.Events&EventRead != 0
}

func (e *Event) Writable() bool {
	return e.Events&EventWrite != 0
}

func (e *Event) Read(p []byte) (int, error) {
	return syscall.Read(e.Fd, p)
}

func (e *Event) Write(p []byte) (int, error) {
	return syscall.Write(e.Fd, p)
}

type Callback func(event *Event)

// 默认使用ET(EdgeTriggered)模式
// 读事件则需要全部读取完
// 写事件:
//	LT模式下,需要则添加,不需要则删除EventWrite
//	ET模式下,当写空间不足时,添加EventWrite事件即可
type poller interface {
	IsSupportET() bool
	Open() error
	Close() error

	Wakeup() error
	Wait(cb Callback) error

	// 注册fd,并监听读事件
	Add(fd fd_t) error
	// 注销fd,并删除读写事件
	Del(fd fd_t) error
	// 添加或删除写监听
	ModifyWrite(fd fd_t, add bool) error
}
