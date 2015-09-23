package common

import (
	"strconv"
	"strings"
)

func EncodeRespStringTo(data []string, str string) []string {
	return []string{"$", strconv.Itoa(len(str)), " ", str}
}

func EncodeRespInt64To(data []string, val int64) []string {
	return []string{":", strconv.FormatInt(val, 10)}
}

func EncodeRespString(data string) string {
	output := []string{"$", strconv.Itoa(len(data)), " ", data}
	return strings.Join(output, "")
}

func EncodeRespInt64(val int64) string {
	return ":" + strconv.FormatInt(val, 10)
}
