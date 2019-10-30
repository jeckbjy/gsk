package alog

type Formatter interface {
	ID() int // 唯一ID,用于服用格式化结果
	Name() string
	Parse(layout string) error
	Format(entry *Entry) ([]byte, error)
}

var id int

func NewFormatterID() int {
	id++
	return id
}

func NewFormatter(name string) Formatter {
	switch name {
	case "text":
		return &TextFormatter{id: NewFormatterID()}
	case "json":
		return &JsonFormatter{id: NewFormatterID()}
	default:
		return nil
	}
}
