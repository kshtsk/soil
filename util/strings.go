package util

import (
	"strings"
)

func SplitLast(s, sep string) (last string) {
	if i := strings.LastIndex(s, sep); i >= 0 {
		return s[i+len(sep):]
	}
	return s
}
