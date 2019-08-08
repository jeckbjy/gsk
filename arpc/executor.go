package arpc

import "sync"

func NewExecutor() IExecutor {
	return &Executor{strategy: DefaultHashStrategy}
}

type HashStrategy func(req IPacket) int

// 默认线程模型,单线程执行
func DefaultHashStrategy(_ IPacket) int {
	return 0
}

// 可以有多种线程模型,默认单线程
// 1:最简单的单线程
// 2:多个线程,每个线程一个消息队列,根据消息ID，hash到不同的线程中
// 3:多个工作线程(极限每个消息起一个协程),一个消息队列,不关心消息执行顺序
type Executor struct {
	workers     []*QWorker
	middlewares []Middleware
	strategy    HashStrategy
	mux         sync.Mutex // use spinlock
}

func (e *Executor) Use(middleware ...Middleware) {
	e.middlewares = append(e.middlewares, middleware...)
}

func (e *Executor) Handle(ctx IContext) {
	index := e.strategy(ctx.Request())
	worker := e.GetWorker(index)
	worker.Post(ctx)
}

func (e *Executor) GetWorker(index int) *QWorker {
	// TODO: 预创建，不使用锁,或者使用spinlock?
	var worker *QWorker

	e.mux.Lock()
	if index >= len(e.workers) {
		worker = &QWorker{}
		worker.Init(e)
		worker.Run()
		e.workers = append(e.workers, worker)
	} else {
		worker = e.workers[index]
	}
	e.mux.Unlock()
	return worker
}

func (e *Executor) apply(ctx IContext) error {
	h := ctx.Handler()
	// TODO: middleware基本是静态创建的,有没有办法减少这次动态创建?
	for i := len(e.middlewares) - 1; i >= 0; i-- {
		h = e.middlewares[i](h)
	}

	return h(ctx)
}

// QWorker 工作队列
type QWorker struct {
	quit  bool // 是否需要退出
	exec  *Executor
	mux   sync.Mutex
	cond  *sync.Cond
	queue QList
}

func (w *QWorker) Init(exec *Executor) {
	w.cond = sync.NewCond(&w.mux)
	w.exec = exec
	w.quit = false
}

func (w *QWorker) Post(ctx IContext) {
	w.mux.Lock()
	node := gGlobalQueuePool.Obtain()
	node.ctx = ctx
	w.queue.Push(node)
	w.mux.Unlock()
	w.cond.Signal()
}

func (w *QWorker) Run() {
	var queue QList
	var quit bool
	for {
		quit = false
		w.mux.Lock()

		for !w.quit && w.queue.Empty() {
			w.cond.Wait()
		}
		quit = w.quit
		queue.Swap(&w.queue)

		w.mux.Unlock()

		for !queue.Empty() {
			node := queue.Pop()
			w.exec.apply(node.ctx)
			gGlobalQueuePool.Free(node)
		}

		if quit {
			break
		}
	}
}

func (w *QWorker) Stop() {
	w.mux.Lock()
	w.quit = true
	w.mux.Unlock()
	w.cond.Signal()
}

// QNode QueueNode
type QNode struct {
	prev *QNode
	next *QNode
	ctx  IContext
}

// QList QueueList 双向非循环队列
type QList struct {
	head *QNode
	tail *QNode
}

func (l *QList) Swap(o *QList) {
	*l, *o = *o, *l
}

func (l *QList) Empty() bool {
	return l.head == nil
}

func (l *QList) Push(n *QNode) {
	if l.tail == nil {
		l.head = n
		l.tail = n
	} else {
		n.prev = l.tail
		l.tail.next = n
	}
}

func (l *QList) Pop() *QNode {
	if l.head != nil {
		n := l.head
		l.head = n.next
		n.next = nil
		n.prev = nil
		return n
	}

	return nil
}

var gGlobalQueuePool QPool

// QPoll QueueNode pool
type QPool struct {
	mux  sync.Mutex
	free QList
}

func (p *QPool) Obtain() *QNode {
	p.mux.Lock()
	defer p.mux.Unlock()
	node := p.free.Pop()
	if node == nil {
		node = &QNode{}
	}

	return node
}

func (p *QPool) Free(n *QNode) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.free.Push(n)
}
