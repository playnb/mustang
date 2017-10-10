package until

import (
	"math/rand"
	"time"
)

var _rand *rand.Rand

func init() {
	_rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

//RandInt 随机数函数
func RandInt(n int) int {
	return _rand.Intn(n)
}

//RandInt64 随机数函数
func RandInt64(n int64) int64 {
	return _rand.Int63n(n)
}
