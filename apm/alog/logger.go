package alog

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	statusNone = 0
	statusRun  = 1
	statusStop = 2
)

func New() *Logger {
	l := &Logger{}
	l.init()
	return l
}

// Logger 异步Log,溢出会丢弃
type Logger struct {
	mux       sync.Mutex
	status    int
	channels  []Channel
	pool      sync.Pool
	cond      Cond
	queue     Queue
	max       int
	level     Level
	formatter Formatter
}

func (l *Logger) AddChannel(c Channel) {
	l.mux.Lock()
	defer l.mux.Unlock()
	c.SetLogger(l)
	l.channels = append(l.channels, c)
}

func (l *Logger) SetLevel(lv Level) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.level = lv
}

func (l *Logger) SetFormatter(f Formatter) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.formatter = f
}

func (l *Logger) init() {
	l.pool.New = func() interface{} {
		return &Entry{}
	}
	l.max = math.MinInt32
	l.level = LevelTrace
	l.cond.Init()
}

func (l *Logger) Start() {
	l.mux.Lock()
	if l.status == statusNone {
		l.status = statusRun
	}
	l.mux.Unlock()
	go l.Run()
}

func (l *Logger) Stop() {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.status != statusRun {
		return
	}
	l.status = statusStop
	l.cond.Signal()
}

func (l *Logger) Run() {
	l.mux.Lock()
	if len(l.channels) == 0 {
		l.channels = append(l.channels, NewTerminal())
	}
	if l.formatter == nil {
		l.formatter, _ = NewTextFormatter("")
	}
	l.mux.Unlock()

	for {
		quit := false
		queue := Queue{}
		l.cond.Lock()
		for l.status != statusStop && l.queue.Empty() {
			l.cond.Wait()
		}
		l.queue.Swap(&queue)
		quit = l.status == statusStop
		l.cond.Unlock()

		for {
			e := queue.Pop()
			if e == nil {
				break
			}
			e.formatter = l.formatter
			// process
			l.mux.Lock()
			for _, channel := range l.channels {
				lv := channel.GetLevel()
				if lv != LevelInherit && e.Level > lv {
					continue
				}

				channel.Write(e)
			}
			l.mux.Unlock()
			l.pool.Put(e)
		}

		if quit {
			break
		}
	}
}

func (l *Logger) WithFields(fields map[string]string) *Builder {
	return &Builder{logger: l, fields: fields}
}

func (l *Logger) Push(e *Entry) {
	needRun := false
	l.cond.Lock()
	l.queue.Push(e)
	if l.status == statusNone {
		l.status = statusRun
		needRun = true
	}
	l.cond.Unlock()
	l.cond.Signal()

	if needRun {
		go l.Run()
	}
}

func (l *Logger) Write(lv Level, fields map[string]string, skipFrames int, text string) {
	e := l.pool.Get().(*Entry)
	e.Frame = getFrame(skipFrames)
	e.formatter = l.formatter
	e.Time = time.Now()
	e.Level = lv
	e.Text = text
	e.Fields = fields
	l.Push(e)
}

func (l *Logger) Log(lv Level, args ...interface{}) {
	l.Write(lv, nil, 1, fmt.Sprint(args...))
}

func (l *Logger) Logf(lv Level, format string, args ...interface{}) {
	l.Write(lv, nil, 1, fmt.Sprintf(format, args...))
}

func (l *Logger) Trace(args ...interface{}) {
	l.Log(LevelTrace, args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.Log(LevelDebug, args...)
}

func (l *Logger) Print(args ...interface{}) {
	l.Log(LevelPrint, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Log(LevelInfo, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.Log(LevelWarn, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Log(LevelError, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Log(LevelFatal, args...)
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Logf(LevelTrace, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logf(LevelDebug, format, args...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.Logf(LevelPrint, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logf(LevelInfo, format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logf(LevelWarn, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logf(LevelError, format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logf(LevelFatal, format, args...)
}
