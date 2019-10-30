package base

import (
	"sync"

	"github.com/jeckbjy/gsk/exec"
)

func NewWorker() *Worker {
	return &Worker{}
}

type Worker struct {
	queue Queue
	mux   *sync.Mutex
	cond  *sync.Cond
	wg    *sync.WaitGroup
	quit  bool
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	w.mux = &sync.Mutex{}
	w.cond = sync.NewCond(w.mux)
	w.quit = false
	w.wg = wg
	go w.run()
}

func (w *Worker) Stop() {
	w.mux.Lock()
	w.quit = true
	w.mux.Unlock()
	w.cond.Signal()
}

func (w *Worker) Post(task exec.Task) {
	w.mux.Lock()
	w.queue.Push(task)
	w.mux.Unlock()
	w.cond.Signal()
}

func (w *Worker) run() {
	w.wg.Add(1)
	defer w.wg.Done()

	var queue Queue
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

		queue.Process()

		if quit {
			break
		}
	}
}
