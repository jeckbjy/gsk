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

// Logger 支持同步或者异步日志输出,默认使用同步控制台输出
type Logger struct {
	mux         sync.Mutex
	status      int
	channels    []Channel
	pool        sync.Pool
	cond        Cond
	queue       Queue
	max         int
	level       Level
	fields      map[string]string // 可以设置一些全局的数据,比如env,service,node等
	formatter   Formatter         // 默认编码格式
	synchronous bool              // 是否同步
}

func (l *Logger) AddChannel(c Channel) {
	l.mux.Lock()
	c.SetLogger(l)
	l.channels = append(l.channels, c)
	l.mux.Unlock()
}

func (l *Logger) Level() Level {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.level
}

func (l *Logger) SetLevel(lv Level) {
	l.mux.Lock()
	l.level = lv
	l.mux.Unlock()
}

func (l *Logger) Formatter() Formatter {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.formatter
}

func (l *Logger) SetFormatter(f Formatter) {
	l.mux.Lock()
	l.formatter = f
	l.mux.Unlock()
}

func (l *Logger) Max() int {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.max
}

func (l *Logger) SetMax(max int) {
	l.mux.Lock()
	l.max = max
	l.mux.Unlock()
}

func (l *Logger) Sync() bool {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.synchronous
}

func (l *Logger) SetSync(s bool) {
	l.mux.Lock()
	l.synchronous = s
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
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.fields != nil {
		if v, ok := l.fields[key]; ok {
			return v
		}
	}

	return ""
}

func (l *Logger) WithFields(fields map[string]string) *Builder {
	return &Builder{logger: l, fields: fields}
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
			// process
			l.mux.Lock()
			l.process(e)
			l.mux.Unlock()
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

func (l *Logger) process(e *Entry) {
	for _, channel := range l.channels {
		lv := channel.GetLevel()
		if e.Level < lv {
			continue
		}

		channel.Write(e)
	}

	// 注意:这里使用完就回收了
	// 如果希望post另外的协程中处理,需要单独拷贝一份,或者直接保存序列化后的结果
	l.pool.Put(e)
}

func (l *Logger) Push(e *Entry) {
	l.cond.Lock()
	canSignal := false
	if l.queue.Len() < l.max {
		l.queue.Push(e)
		canSignal = true
	}
	l.cond.Unlock()
	if canSignal {
		l.cond.Signal()
	}
}

func (l *Logger) Write(lv Level, fields map[string]string, skipFrames int, text string) {
	l.mux.Lock()
	if l.status == statusNone {
		// 初始化设置默认参数
		l.status = statusRun
		if len(l.channels) == 0 {
			// 从来没有设置过channel,默认使用同步控制台输出
			l.synchronous = true
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
		// 异步
		if !l.synchronous {
			go l.Run()
		}
	}

	formatter := l.formatter
	synchronous := l.synchronous
	l.mux.Unlock()

	e := l.pool.Get().(*Entry)
	e.Reset()
	e.Frame = getFrame(skipFrames)
	e.formatter = formatter
	e.Time = time.Now()
	e.Level = lv
	e.Text = text
	e.Fields = fields
	e.logger = l

	if synchronous {
		l.process(e)
	} else {
		l.Push(e)
	}
}

// TODO: fmt.Sprint 会紧凑的合并到一起,期望能自动添加分隔符
func (l *Logger) Log(lv Level, args ...interface{}) {
	if lv >= l.level {
		l.Write(lv, nil, 3, fmt.Sprint(args...))
	}
}

func (l *Logger) Logf(lv Level, format string, args ...interface{}) {
	if lv >= l.level {
		l.Write(lv, nil, 3, fmt.Sprintf(format, args...))
	}
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
