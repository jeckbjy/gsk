package luhn

import (
	"errors"
	"fmt"
)

var ErrNotSupport = errors.New("not support type, just support string,uint64")

// https://www.geeksforgeeks.org/luhn-algorithm/
// 支持string和uint64,int,uint
func Check(v interface{}) bool {
	switch v.(type) {
	case int:
		return checkInt(uint64(v.(int)))
	case uint:
		return checkInt(uint64(v.(uint)))
	case uint64:
		return checkInt(v.(uint64))
	case string:
		s := v.(string)
		if len(s) < 2 {
			return false
		}
		l := s[len(s)-1]
		t := s[:len(s)-1]
		c, err := checksumStr(t)
		if err != nil {
			return false
		}
		return l == c
	default:
		return false
	}
}

func GenerateInt(x uint64) uint64 {
	l := checksumInt(x)
	return x*10 + uint64(l)
}

func GenerateStr(s string) string {
	l, err := checksumStr(s)
	if err != nil {
		return ""
	}

	return s + string(l)
}

func Generate(v interface{}) (interface{}, error) {
	switch v.(type) {
	case int:
		return GenerateInt(uint64(v.(int))), nil
	case uint:
		return GenerateInt(uint64(v.(uint))), nil
	case uint64:
		return GenerateInt(v.(uint64)), nil
	case string:
		s := v.(string)
		l, err := checksumStr(s)
		if err != nil {
			return v, err
		}

		return s + string(l), nil
	default:
		return nil, ErrNotSupport
	}
}

func checkInt(x uint64) bool {
	l := x % 10
	t := x / 10
	return checksumInt(t) == byte(l)
}

func checksumInt(number uint64) byte {
	l := uint64(0)
	even := true
	for number > 0 {
		cur := number % 10
		if even {
			cur *= 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		l += cur
		number /= 10
		even = !even
	}

	result := l % 10
	if result > 0 {
		result = 10 - result
	}

	return byte(result)
}

func checksumStr(s string) (byte, error) {
	l := 0
	even := true
	for i := len(s) - 1; i >= 0; i-- {
		b := s[i]
		if b < '0' || b > '9' {
			return 0, fmt.Errorf("is not digit")
		}

		cur := b - '0'
		if even {
			cur *= 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		l += int(cur)
		even = !even
	}

	result := l % 10
	if result > 0 {
		result = 10 - result
	}

	return byte(result) + '0', nil
}
