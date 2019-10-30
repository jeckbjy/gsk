package exec

import (
	"errors"
)

var (
	ErrAlreadyStop = errors.New("already stop")
	//ErrBadHashIndex = errors.New("bad hash index")
)

type Task interface {
	Run() error
}

// 线程/协程模型
//	1:单线程
//	2:根据ID hash到某个线程
//	3:线程池,不超过最大线程数
//  4:每个消息起一个go routine?
type Executor interface {
	Handle(task Task) error
	Stop() error
	Wait()
}
