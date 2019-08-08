package reflects

import (
	"runtime"
	"strings"
)

// 返回函数名,不包含包名
func FuncName(pc uintptr) string {
	fullname := runtime.FuncForPC(pc).Name()
	index := strings.LastIndexByte(fullname, '.')
	if index == -1 {
		return ""
	}

	return fullname[index+1:]
}
