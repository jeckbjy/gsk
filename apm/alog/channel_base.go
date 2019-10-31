package alog

type BaseChannel struct {
	logger    *Logger
	level     Level
	formatter Formatter
}

func (c *BaseChannel) SetLogger(l *Logger) {
	c.logger = l
}

func (c *BaseChannel) SetLevel(lv Level) {
	c.level = lv
}

func (c *BaseChannel) GetLevel() Level {
	return c.level
}

func (c *BaseChannel) SetFormatter(formatter Formatter) {
	c.formatter = formatter
}

func (c *BaseChannel) GetFormatter() Formatter {
	return c.formatter
}

func (c *BaseChannel) SetProperty(key string, value string) error {
	switch key {
	case "level":
		level, err := ParseLevel(value)
		if err != nil {
			return err
		}
		c.level = level
	case "formatter":

	}
	return nil
}

func (c *BaseChannel) GetProperty(key string) string {
	return ""
}
