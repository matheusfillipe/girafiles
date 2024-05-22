package api

import (
	"math"

	"github.com/jxskiss/base62"
)

const (
	MIN_LENGTH = 2
)

func getOffset() int64 {
	return int64(math.Pow(62, MIN_LENGTH-1))
}

// / Convert index to equivalent alphanumeric string
func IdxToString(i int64) string {
	offset := getOffset()
	output := string(base62.FormatInt(i + int64(offset)))
	return output
}

// / Convert alphanumeric string to equivalent index
func StringToIdx(s string) (int64, error) {
	offset := getOffset()
	output, error := base62.ParseInt([]byte(s))
	if error != nil {
		return 0, error
	}
	return output - int64(offset), nil
}
