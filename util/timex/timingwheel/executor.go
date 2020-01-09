package timingwheel

import "sync"

const (
	statusNone = 0 // 尚未运行
	statusRun  = 1 // 运行中
	statusExit = 2 // 已经退出
)

// 单独一个协程用于执行定时器,不保证一定是有序,通常是有序执行,但是Stop时可能会乱序
// 一旦进入executor,timer就不能撤销了,虽然timer并没有执行,这样有助于简化锁的使用
type executor struct {
	working *bucket
	waiting *bucket
	mux     sync.Mutex
	cond    *sync.Cond
	status  int
}

func (e *executor) Start() {
	e.status = statusRun
	e.cond = sync.NewCond(&e.mux)
	e.working = newBucket()
	e.waiting = newBucket()
	go e.Run()
}

func (e *executor) Stop() {
	e.mux.Lock()
	e.status = statusExit
	pendings := *e.waiting
	e.waiting.Reset()
	e.mux.Unlock()
	e.cond.Signal()
	process(&pendings)
}

// 投递到另外一个线程执行
func (e *executor) Post(p *bucket) {
	if p.Empty() {
		return
	}

	e.mux.Lock()
	e.waiting.merge(p)
	e.mux.Unlock()
	e.cond.Signal()
}

func (e *executor) Run() {
	for {
		e.mux.Lock()
		for e.status != statusRun || e.waiting.Empty() {
			e.cond.Wait()
		}
		exit := e.status == statusExit

		// swap list
		pendings := e.waiting
		e.waiting, e.working = e.working, e.waiting

		e.mux.Unlock()

		process(pendings)

		if exit {
			break
		}
	}
}

func process(pendings *bucket) {
	for iter := pendings.Front(); iter != nil; {
		timer := iter
		iter = iter.next
		pendings.Remove(timer)
		// 调用前已经被删除了,可以再次被调用
		timer.task()
	}
}
