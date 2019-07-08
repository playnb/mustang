package util

import (
	"testing"
)

func TestRandBetween(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := RandInt64(100)
		b := RandInt64(100)
		c := RandBetween(a, b)
		n := ((c >= a) && (c <= b)) || ((c >= b) && (c <= a))
		//fmt.Printf("%d  %d ==> %d  %v\n", a, b, c, t)
		if !n {
			t.Error("失败")
		}
	}
}
