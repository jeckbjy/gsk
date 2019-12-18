package internal

import "errors"

const maxEventNum = 1024

const (
	EventRead  = 0x01
	EventWrite = 0x02
	EventError = 0x08
)

var (
	ErrNotSupport = errors.New("not support")
)

type Event struct {
	poll   Poller
	fd     FD
	events int
}

func (e *Event) Fd() FD {
	return e.fd
}

func (e *Event) HasError() bool {
	return e.events&EventError != 0
}

func (e *Event) Readable() bool {
	return e.events&EventRead != 0
}

func (e *Event) Writable() bool {
	return e.events&EventWrite != 0
}

func (e *Event) Delete() error {
	return e.poll.Delete(e.fd)
}

type Callback func(event *Event)

func New() (Poller, error) {
	poll := newPoller()
	if err := poll.Open(); err != nil {
		return nil, err
	}

	return poll, nil
}

// 默认使用ET(EdgeTriggered)模式
// 读事件则需要全部读取完
// 写事件:
//	LT模式下,需要则添加,不需要则删除EventWrite
//	ET模式下,当写空间不足时,添加EventWrite事件即可
type Poller interface {
	IsSupportET() bool
	Open() error
	Close() error

	Wakeup() error
	Wait(cb Callback) error

	// 注册fd,并监听读事件
	// TODO:如何添加关联数据,可以更加高效的处理回调
	Add(fd FD) error
	// 注销fd,并删除读写事件
	Delete(fd FD) error
	// 添加或删除写监听
	ModifyWrite(fd FD, add bool) error
}
