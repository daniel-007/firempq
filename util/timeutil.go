package util

import (
	"time"
)

func Uts() int64 {
	return time.Now().UnixNano() / 1000000
}
