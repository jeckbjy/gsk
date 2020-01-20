package alog

const DefaultTextLayout = "%D [%L] [%F] %m"

func NewTextFormatter(layout string) (Formatter, error) {
	if layout == "" {
		layout = DefaultTextLayout
	}
	f := &TextFormatter{id: NewFormatterID()}
	if err := f.Parse(layout); err != nil {
		return nil, err
	}

	return f, nil
}

type TextFormatter struct {
	id     int
	layout *Layout
}

func (f *TextFormatter) ID() int {
	return f.id
}

func (f *TextFormatter) Name() string {
	return "text"
}

func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	data := f.layout.Format(entry)
	return []byte(data), nil
}

func (f *TextFormatter) Parse(layout string) error {
	l, err := NewLayout(layout)
	if err != nil {
		return err
	}
	f.layout = l
	return nil
}
