package util

import (
	"cell/common/mustang/log"
	"encoding/gob"
	"math/rand"
	"os"
	"sync"
	"time"
)

var _rand *rand.Rand

func init() {
	_rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandBetween(a int64, b int64) int64 {
	n := int64(0)
	if a < b {
		n = b - a
	} else if a > b {
		n = a - b
	} else {
		return a
	}
	n = RandInt64(n)
	if a < b {
		return n + a
	}
	return n + b
}

//RandInt 随机数函数
func RandInt(n int) int {
	if n == 0 {
		return 0
	}
	if _rand_nums_ok {
		_rand_lock.Lock()
		_rand_index = _rand_index + 1
		num := _rand_nums[_rand_index%_rand_index_max]
		_rand_lock.Unlock()
		return int(num) % n
	} else {
		return _rand.Intn(n)
	}
}

//RandInt64 随机数函数
func RandInt64(n int64) int64 {
	if n == 0 {
		return 0
	}
	if _rand_nums_ok {
		_rand_lock.Lock()
		_rand_index = _rand_index + 1
		num := _rand_nums[_rand_index%_rand_index_max]
		_rand_lock.Unlock()
		return int64(num) % n
	} else {
		return _rand.Int63n(n)
	}
}

//===============================================================================
var _rand_nums []int32
var _rand_nums_ok bool
var _rand_index int
var _rand_index_max int
var _rand_lock sync.Mutex

func LoadRandNumFromFile(fileName string) {
	file, _ := os.Open("rand.data")
	dec := gob.NewDecoder(file)
	err := dec.Decode(&_rand_nums)
	if err != nil {
		_rand_nums_ok = false
		_rand_index_max = 0
		log.Error("随机数读取失败 %v", err)
	} else if len(_rand_nums) > 1000 {
		_rand_nums_ok = true
		_rand_index_max = len(_rand_nums)
	}
	file.Close()
}
