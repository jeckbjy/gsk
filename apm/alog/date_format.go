package alog

import (
	"bytes"
	"fmt"
	"time"
)

const (
	df_literal dfType = iota // 字符
	df_d1                    // d
	df_d2                    // dd
	df_d3                    // ddd
	df_d4                    // dddd
	df_h1                    // h
	df_h2                    // hh
	df_H1                    // H
	df_H2                    // HH
	df_m1                    // m
	df_m2                    // mm
	df_M1                    // M
	df_M2                    // MM
	df_M3                    // MMM
	df_M4                    // MMMM
	df_s1                    // s
	df_s2                    // ss
	df_t1                    // t
	df_t2                    // tt
	df_y1                    // y
	df_y2                    // yy
	df_y3                    // yyy
	df_y4                    // yyyy
	df_z1                    // z
	df_z2                    // zz
	df_z3                    // zzz
)

type dfType int

var keyToType = map[string]dfType{
	"d":    df_d1,
	"dd":   df_d2,
	"ddd":  df_d3,
	"dddd": df_d4,
	"h":    df_h1,
	"hh":   df_h2,
	"H":    df_H1,
	"HH":   df_H2,
	"m":    df_m1,
	"mm":   df_m2,
	"M":    df_M1,
	"MM":   df_M2,
	"MMM":  df_M3,
	"MMMM": df_M4,
	"s":    df_s1,
	"ss":   df_s2,
	"t":    df_t1,
	"tt":   df_t2,
	"y":    df_y1,
	"yy":   df_y2,
	"yyy":  df_y3,
	"yyyy": df_y4,
	"z":    df_z1,
	"zz":   df_z2,
	"zzz":  df_z3,
}

func parseType(key string) dfType {
	if t, ok := keyToType[key]; ok {
		return t
	}

	return df_literal
}

type dfToken struct {
	Type dfType
	Data string
}

// https://docs.microsoft.com/en-us/dotnet/standard/base-types/custom-date-and-time-format-strings
// yyyy-MM-ddTHH:mm:ss
type DateFormat struct {
	Tokens []dfToken
}

func (d *DateFormat) Parse(layout string) {
	for idx := 0; idx < len(layout); {
		p := layout[idx]
		beg := idx
		for ; idx < len(layout); idx++ {
			if c := layout[idx]; c != p {
				key := layout[beg:idx]
				d.Tokens = append(d.Tokens, dfToken{Type: parseType(key), Data: key})
				break
			}
		}

		if idx == len(layout) {
			key := layout[beg:]
			d.Tokens = append(d.Tokens, dfToken{Type: parseType(key), Data: key})
		}
	}
}

func (d *DateFormat) Format(t time.Time) string {
	b := bytes.Buffer{}
	for _, token := range d.Tokens {
		switch token.Type {
		case df_literal:
			b.WriteString(token.Data)
		case df_d1:
			dfWrite(&b, "%d", t.Day())
		case df_d2:
			dfWrite(&b, "%2d", t.Day())
		case df_d3:
			dfWrite(&b, "%s", t.Weekday().String()[:3])
		case df_d4:
			dfWrite(&b, "%s", t.Weekday().String())
		case df_h1:
			dfWrite(&b, "%d", dfToHour(t.Hour()))
		case df_h2:
			dfWrite(&b, "%02d", dfToHour(t.Hour()))
		case df_H1:
			dfWrite(&b, "%d", t.Hour())
		case df_H2:
			dfWrite(&b, "%02d", t.Hour())
		case df_m1:
			dfWrite(&b, "%d", t.Minute())
		case df_m2:
			dfWrite(&b, "%02d", t.Minute())
		case df_M1:
			dfWrite(&b, "%d", t.Month())
		case df_M2:
			dfWrite(&b, "%02d", t.Month())
		case df_M3:
			dfWrite(&b, "%s", t.Month().String()[:3])
		case df_M4:
			dfWrite(&b, "%s", t.Month().String())
		case df_s1:
			dfWrite(&b, "%d", t.Second())
		case df_s2:
			dfWrite(&b, "%02d", t.Second())
		case df_t1:
			dfWrite(&b, "%s", dfToAMPM(t.Hour())[0])
		case df_t2:
			dfWrite(&b, "%s", dfToAMPM(t.Hour()))
		case df_y1:
			dfWrite(&b, "%d", t.Year()%10)
		case df_y2:
			dfWrite(&b, "%02d", t.Year()%100)
		case df_y3:
			dfWrite(&b, "%03d", t.Year()%1000)
		case df_y4:
			dfWrite(&b, "%04d", t.Year())
		case df_z1:
			dfWrite(&b, "%s", t.Format("Z07"))
		case df_z2:
			dfWrite(&b, "%s", t.Format("-07"))
		case df_z3:
			dfWrite(&b, "%s", t.Format("-07:00"))
		}
	}

	return b.String()
}

func dfWrite(b *bytes.Buffer, format string, value interface{}) {
	text := fmt.Sprintf(format, value)
	b.WriteString(text)
}

func dfToHour(h int) int {
	if h < 12 {
		return h
	}

	return h - 12
}

func dfToAMPM(h int) string {
	if h < 12 {
		return "AM"
	}

	return "PM"
}
