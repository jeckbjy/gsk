package alog

import (
	"bytes"
	"errors"
	"fmt"
	"time"
	"unicode"
)

func NewLayout(format string) (*Layout, error) {
	l := &Layout{}
	if err := l.Parse(format); err != nil {
		return nil, err
	}

	return l, nil
}

// 解析规则:%{width}key[params],例如,%t,%2p,%c[yyyy-MM-ddTHH:mm:ss]
// %% %
// %m log message
// %L full log level,Info, Warn
// %f function name e.g. model.Get
// %F function name and line number e.g. model.Get:121
// %s short file name and line number: d.go:23
// %S full file name and line number: /a/b/c/d.go:23
//
// %t short time format e.g. "15:04:05"
// %T long time format e.g. "15:04:05.999"
// %d short date format e.g. "2006/01/02"
// %D long date format e.g. "2006/01/02 15:04:05"
// %r RFC1123  date format e.g. Mon, 02 Jan 2006 15:04:05 MST
// %R RFC1123Z date format e.g. Mon, 02 Jan 2006 15:04:05 -0700
// %o RFC3339  date format e.g. 2006-01-02T15:04:05Z07:00
// %{x}[key] additional fields, e.g. %[env] means find field value by 'env' if exits
// %c[format] custom datetime,d,h,H,m,M,s,t,y,:,/
//
// DateFormat: https://docs.microsoft.com/en-us/dotnet/standard/base-types/custom-date-and-time-format-strings

type Layout struct {
	format  string
	actions []*_Action
}

// %{width}key[params]
type _Action struct {
	Key     byte        // %xx
	Prepend string      // xx%
	Width   string      // %{width}
	Param   string      // %[param],原始参数
	Data    interface{} // 计算处理后数据
}

func (l *Layout) Parse(format string) error {
	actions := make([]*_Action, 0, 8)

	end := len(format)
	cur := 0
	for cur < end {
		act := &_Action{}
		// parse prepend
		for beg := cur; cur < end; cur++ {
			if format[cur] == '%' {
				if beg < cur {
					act.Prepend = format[beg:cur]
				}
				break
			}
		}

		// check end
		if cur == end {
			if act.Prepend != "" {
				actions = append(actions, act)
			}
			break
		}

		// parse action
		cur++ // ignore %
		// parse width
		if cur >= end {
			return errors.New("unknown token")
		}

		if format[cur] == '-' || unicode.IsDigit(rune(format[cur])) {
			beg := cur
			cur++
			for ; cur < end; cur++ {
				if !unicode.IsDigit(rune(format[cur])) {
					act.Width = format[beg:cur]
					break
				}
			}
		}

		if cur >= end {
			return errors.New("unknown token")
		}

		// parse key
		if format[cur] == '[' {
			act.Key = 'x'
		} else {
			act.Key = format[cur]
			cur++
		}

		// parse params [xxx]
		if cur < end && format[cur] == '[' {
			cur++
			findClose := false
			for beg := cur; cur < end; cur++ {
				if format[cur] == ']' {
					act.Param = format[beg:cur]
					cur++
					findClose = true
					break
				}
			}
			if findClose {
				return fmt.Errorf("not find close bracket")
			}
		}
		actions = append(actions, act)
	}

	l.actions = actions
	return nil
}

func (l *Layout) Format(msg *Entry) string {
	lb := _LayoutBuilder{}
	for _, act := range l.actions {
		lb.Put(act.Prepend, "")
		switch act.Key {
		case '%':
			lb.Put("%", "")
		case 'm': // message
			lb.Put(msg.Text, act.Width)
		case 'L': // level
			lb.Put(msg.Level.String(), act.Width)
		case 'f':
			lb.Put(msg.FuncName(), act.Width)
		case 'F':
			lb.Put(msg.FuncLine(), act.Width)
		case 's':
			lb.Put(msg.Source(true), act.Width)
		case 'S':
			lb.Put(msg.Source(false), act.Width)
		case 't':
			lb.Put(msg.Time.Format("15:04:05"), act.Width)
		case 'T':
			lb.Put(msg.Time.Format("15:04:05.000"), act.Width)
		case 'd':
			lb.Put(msg.Time.Format("2006/01/02"), act.Width)
		case 'D':
			lb.Put(msg.Time.Format("2006/01/02 15:04:05"), act.Width)
		case 'r':
			lb.Put(msg.Time.Format(time.RFC1123), act.Width)
		case 'R':
			lb.Put(msg.Time.Format(time.RFC1123Z), act.Width)
		case 'o':
			lb.Put(msg.Time.Format(time.RFC3339), act.Width)
		case 'c':
			if text := l.formatDatetime(act, msg.Time); text != "" {
				lb.Put(text, act.Width)
			}
		case 'x':
			if msg.Fields != nil && act.Param != "" {
				if val := msg.GetField(act.Param); val != "" {
					lb.Put(val, act.Width)
				}
			}
		default:

		}
	}

	lb.AddNewLine()
	return lb.String()
}

func (l *Layout) formatDatetime(act *_Action, t time.Time) string {
	var dt *DateFormat
	if d, ok := act.Data.(*DateFormat); ok {
		dt = d
	} else {
		dt = &DateFormat{}
		dt.Parse(act.Param)
		act.Data = dt
	}
	return ""
}

func toUpper(x byte) byte {
	if x <= 'Z' {
		return x
	}

	x -= 'a' - 'A'
	return x
}

type _LayoutBuilder struct {
	builder bytes.Buffer
	last    byte
}

func (l *_LayoutBuilder) Put(data interface{}, width string) {
	var text string
	if str, ok := data.(string); ok {
		text = str
	} else {
		text = fmt.Sprintf("%+v", data)
	}

	if len(text) == 0 {
		return
	}

	if len(width) > 0 {
		text = fmt.Sprintf("%"+width+"s", text)
	}

	l.last = text[len(text)-1]
	l.builder.WriteString(text)
}

// 添加换行
func (l *_LayoutBuilder) AddNewLine() {
	if l.builder.Len() > 0 && l.last != '\n' {
		l.builder.WriteByte('\n')
	}
}

func (l *_LayoutBuilder) String() string {
	return l.builder.String()
}
