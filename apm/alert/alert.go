package alert

var alerts = map[string]Alert{}

type EventType string

const (
	TextEvent EventType = "text"
)

type Event struct {
	Tag   string    // 用于标识使用哪个组
	Type  EventType // 消息类型
	Title string    // 标题,大部分没有
	Text  string    // 内容,必须有
	Meta  map[string]interface{}
}

type Alert interface {
	Name() string
	Send(event *Event) error
}

// 需要启动时注册,无锁,非线程安全
func Register(a Alert) {
	alerts[a.Name()] = a
}

// 发送消息
func Send(event *Event) error {
	var err error
	for _, v := range alerts {
		if e := v.Send(event); e != nil {
			err = e
		}
	}

	return err
}
