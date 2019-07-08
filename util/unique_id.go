package util

import (
	"sync"
	"sync/atomic"
	"time"
)

var _gen *UniqueIDGeneration

func init() {
	_gen = &UniqueIDGeneration{}
	_gen.Init()
}

func NextUniqueID() uint64 {
	return _gen.Generate()
}

type UniqueIDGeneration struct {
	_sequenceTimestamp uint64
	_sequenceId        uint64
	_baseTime          int64

	_lock sync.Mutex
}

func (uid *UniqueIDGeneration) Init() {
	uid._sequenceTimestamp = uint64(time.Now().UnixNano() / 1000)
	uid._sequenceId = 0
	uid._baseTime = time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local).UnixNano() / 1000
}

//TODO 这个Lock应该不是必须的，先锁着，脑子清楚了再说
func (uid *UniqueIDGeneration) Generate() uint64 {
	uid._lock.Lock()
	defer uid._lock.Unlock()
	nowTime := uint64(time.Now().UnixNano() / 1000)
	if uid._sequenceTimestamp == 0 {
		uid._sequenceTimestamp = nowTime
	}

	i := atomic.AddUint64(&uid._sequenceId, 1) % 1000

	if i == 0 && uid._sequenceTimestamp == nowTime {
		time.Sleep(time.Microsecond)
		nowTime = uint64(time.Now().UnixNano() / 1000)
		i = 0
		//fmt.Printf("nowTime:%d _sequenceId:%d \n", nowTime, uid._sequenceId)
	}

	uid._sequenceTimestamp = nowTime
	token := nowTime
	token = token - uint64(uid._baseTime)
	token = uint64(token*1000) + (uid._sequenceId)
	uid._sequenceTimestamp = nowTime

	return token
}
