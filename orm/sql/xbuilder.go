package sql

import (
	"fmt"
	"strings"
)

const (
	spNone  = 0
	spBlank = ' '
	spComma = ','
)

// 会自动添加分隔符(默认空格)
type sqlBuilder struct {
	builder strings.Builder
	last    byte
}

func (b *sqlBuilder) Init() {
	b.builder.Grow(128)
}

func (b *sqlBuilder) Len() int {
	return b.builder.Len()
}

func (b *sqlBuilder) String() string {
	return b.builder.String()
}

func (b *sqlBuilder) Separate(separator byte) {
	if separator != 0 && b.builder.Len() > 0 && b.last != separator {
		b.builder.WriteByte(separator)
		b.last = separator
	}
}

func (b *sqlBuilder) Write(separator byte, format string, args ...interface{}) {
	b.Separate(separator)
	var text string
	if len(args) > 0 {
		text = fmt.Sprintf(format, args...)
	} else {
		text = format
	}
	b.builder.WriteString(text)
	b.last = text[len(text)-1]
}
