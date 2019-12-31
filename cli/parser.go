package cli

import (
	"errors"
	"runtime"
	"strings"
)

// unix命令行格式解析
// 1:-表示shortcut,可以多个合并,例如,-h 表示help, -czvx，表示-c -z -v -x,如果带参数,则只会设置给最后一个
// 2:--表示全称，例如--help
// 3:后边可以紧跟一个参数,可以使用=连载一起写,也可以空格分隔
// 4:可以重复,相同的则合并成1个处理
type Parser struct {
	params  []string
	options map[string][]string
}

func (p *Parser) Parse(args []string) error {
	for idx := 0; idx < len(args); idx++ {
		token := strings.TrimSpace(args[idx])
		if len(token) == 0 {
			continue
		}

		if token[0] != '-' {
			p.params = append(p.params, token)
			continue
		}

		// parse flag
		if len(token) == 1 {
			// - 只能是最后一个
			if idx < len(args)-1 {
				return errors.New("bad -,not the last")
			} else {
				return p.addOption("-", "")
			}
		}

		var short bool
		var key string
		var val string
		if token[1] == '-' {
			short = false
			key = token[2:]
		} else {
			short = true
			key = token[1:]
		}

		// parse flag value
		if strings.ContainsRune(key, '=') {
			// check has =
			values := strings.SplitN(key, "=", 2)
			key = values[0]
			val = values[1]
		} else if idx+1 < len(args) && args[idx+1][0] != '-' {
			// check next
			idx++
			val = args[idx]
		}

		if short && len(key) > 1 {
			for i := 0; i < len(key)-1; i++ {
				if err := p.addOption(string(key[i]), ""); err != nil {
					return err
				}
			}
		} else {
			if err := p.addOption(key, val); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Parser) addOption(key string, val string) error {
	if p.options == nil {
		p.options = make(map[string][]string)
	}

	if d, ok := p.options[key]; ok {
		if len(val) == 0 || len(d) == 0 {
			// multiple 需要有参数
			return errors.New("multiple options need param")
		}
		p.options[key] = append(d, val)
	} else {
		if len(val) == 0 {
			p.options[key] = nil
		} else {
			p.options[key] = []string{val}
		}
	}

	return nil
}

// 解析命令行参数
// from: https://github.com/mgutz/str.git
func ParseCommandLine(s string) ([]string, error) {
	const (
		InArg = iota
		InArgQuote
		OutOfArg
	)
	currentState := OutOfArg
	currentQuoteChar := "\x00" // to distinguish between ' and " quotations
	// this allows to use "foo'bar"
	currentArg := ""
	var argv []string

	isQuote := func(c string) bool {
		return c == `"` || c == `'`
	}

	isEscape := func(c string) bool {
		return c == `\`
	}

	isWhitespace := func(c string) bool {
		return c == " " || c == "\t"
	}

	L := len(s)
	for i := 0; i < L; i++ {
		c := s[i : i+1]

		//fmt.Printf("c %s state %v arg %s argv %v i %d\n", c, currentState, currentArg, args, i)
		if isQuote(c) {
			switch currentState {
			case OutOfArg:
				currentArg = ""
				fallthrough
			case InArg:
				currentState = InArgQuote
				currentQuoteChar = c

			case InArgQuote:
				if c == currentQuoteChar {
					currentState = InArg
				} else {
					currentArg += c
				}
			}

		} else if isWhitespace(c) {
			switch currentState {
			case InArg:
				argv = append(argv, currentArg)
				currentState = OutOfArg
			case InArgQuote:
				currentArg += c
			case OutOfArg:
				// nothing
			}

		} else if isEscape(c) {
			switch currentState {
			case OutOfArg:
				currentArg = ""
				currentState = InArg
				fallthrough
			case InArg:
				fallthrough
			case InArgQuote:
				if i == L-1 {
					if runtime.GOOS == "windows" {
						// just add \ to end for windows
						currentArg += c
					} else {
						return nil, errors.New("escape character at end string")
					}
				} else {
					if runtime.GOOS == "windows" {
						peek := s[i+1 : i+2]
						if peek != `"` {
							currentArg += c
						}
					} else {
						i++
						c = s[i : i+1]
						currentArg += c
					}
				}
			}
		} else {
			switch currentState {
			case InArg, InArgQuote:
				currentArg += c

			case OutOfArg:
				currentArg = ""
				currentArg += c
				currentState = InArg
			}
		}
	}

	if currentState == InArg {
		argv = append(argv, currentArg)
	} else if currentState == InArgQuote {
		return nil, errors.New("starting quote has no ending quote")
	}

	return argv, nil
}
