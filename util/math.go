package util

import (
	"math"
)

const MIN = 0.000001

func IsEqual(f1, f2 float64) bool {
	return math.Dim(f1, f2) < MIN
}

// math.Abs 针对float64，手动实现一个针对int64
func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
