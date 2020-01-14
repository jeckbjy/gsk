package alog

import (
	"fmt"
	"sync"
	"time"
)

const (
	statusNone = 0
	statusRun  = 1
	statusStop = 2
)

const (
	// 默认最大堆积消息数,超出则丢弃
	DefaultMsgMax = 10000
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
	fields    map[string]string // 可以设置一些全局的数据,比如env,service,node等
	formatter Formatter
}

func (l *Logger) AddChannel(c Channel) {
	l.mux.Lock()
	defer l.mux.Unlock()
	c.SetLogger(l)
	l.channels = append(l.channels, c)
}

func (l *Logger) Level() Level {
	return l.level
}

func (l *Logger) SetLevel(lv Level) {
	l.mux.Lock()
	l.level = lv
	l.mux.Unlock()
}

func (l *Logger) Formatter() Formatter {
	return l.formatter
}

func (l *Logger) SetFormatter(f Formatter) {
	l.mux.Lock()
	l.formatter = f
	l.mux.Unlock()
}

func (l *Logger) Max() int {
	return l.max
}

func (l *Logger) SetMax(max int) {
	l.mux.Lock()
	l.max = max
	l.mux.Unlock()
}

func (l *Logger) AddField(key, value string) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.fields == nil {
		l.fields = make(map[string]string)
	}
	l.fields[key] = value
}

func (l *Logger) GetField(key string) string {
	if l.fields != nil {
		if v, ok := l.fields[key]; ok {
			return v
		}
	}

	return ""
}

func (l *Logger) init() {
	l.pool.New = func() interface{} {
		return &Entry{}
	}
	l.max = DefaultMsgMax
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

	for _, c := range l.channels {
		if err := c.Open(); err != nil {
			fmt.Println(err)
		}
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
	l.mux.Lock()
	for _, c := range l.channels {
		if err := c.Close(); err != nil {
			fmt.Println(err)
		}
	}
	l.mux.Unlock()
}

func (l *Logger) WithFields(fields map[string]string) *Builder {
	return &Builder{logger: l, fields: fields}
}

func (l *Logger) Push(e *Entry) {
	needRun := false
	l.cond.Lock()
	// overflow
	if l.queue.Len() > l.max {
		l.cond.Unlock()
		return
	}

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
	e.logger = l
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
