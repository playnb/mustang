package util

import (
	"fmt"
	"math"
	"sync"
)

var _defaultPool *BuffDataPool

func init() {
	_defaultPool = &BuffDataPool{}
	_defaultPool.Init()
}

func DefaultPool() *BuffDataPool {
	return _defaultPool
}

func MakeBuffData(size int) *BuffData {
	data := &BuffData{}
	data.buffCap = size
	if size > 0 {
		data.data = make([]byte, size)
	}
	data.size = size
	data.index = 0
	data.inUse = true
	return data
}

func MakeBuffDataBySlice(buf []byte, offset int) *BuffData {
	data := MakeBuffData(0)
	if offset > 0 {
		data.data = make([]byte, len(buf)+offset)
		data.index = 0
		data.size = len(data.data)
		data.ChangeIndex(offset)
		copy(data.data[data.index:], buf)
	} else {
		data.data = buf
		data.index = 0
		data.size = len(data.data)
	}
	return data
}

type BuffData struct {
	buffCap int
	inUse   bool
	pool    *BuffDataPool

	size  int
	index int
	data  []byte
}

func (buff *BuffData) GetPayload() []byte {
	if buff != nil {
		if buff.index+buff.size > len(buff.data) {
			//panic(buff)
			fmt.Printf("BuffData 数据异常 %v", buff)
		}
		d := buff.data[buff.index : buff.index+buff.size]
		return d
	}
	return nil
}

func (buff *BuffData) Release() {
	if buff != nil && buff.pool != nil {
		buff.pool.Put(buff)
	}
}

func (buff *BuffData) Size() int {
	if buff != nil {
		return buff.size
	}
	return 0
}

func (buff *BuffData) SetSize(size int) {
	if buff != nil {
		if buff.size < size {
			fmt.Println("BuffData 设置大小出错")
			return
		}
		buff.size = size
	}
}

func (buff *BuffData) Data() []byte {
	if buff != nil {
		return buff.data[:buff.index+buff.size]
	}
	return nil
}

func (buff *BuffData) ChangeIndex(index int) bool {
	if buff != nil {
		if index > 0 && buff.size >= index {
			buff.index += index
			buff.size -= index
			return true
		} else if index < 0 && buff.index >= -index {
			buff.index += index
			buff.size -= index
			return true
		}
	}
	return false
}

func (buff *BuffData) Append(data []byte) {
	if buff != nil {
		copy(buff.data[buff.index+buff.size:], data)
		buff.size = buff.size + len(data)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////
type OneBuffDataPool struct {
	sync.Pool
	cap int
}

func (pool *OneBuffDataPool) String() string {
	return ""
}

/*
type OneBuffDataPool struct {
	//sync.Pool
	cap       int
	New       func() interface{}
	buff      []interface{}
	buffIndex int
	lock      sync.Mutex
}

func (pool *OneBuffDataPool) String() string {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return fmt.Sprintf("缓存:%d Index:%d", len(pool.buff), pool.buffIndex)
}

func (pool *OneBuffDataPool) Get() interface{} {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if pool.buffIndex > 0 && len(pool.buff) > 0 {
		b := pool.buff[pool.buffIndex-1]
		pool.buffIndex--
		return b
	} else {
		return pool.New()
	}
	return nil
}

func (pool *OneBuffDataPool) Put(d interface{}) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if pool.buffIndex < len(pool.buff) {
		pool.buffIndex++
		pool.buff[pool.buffIndex-1] = d
	} else {
		pool.buffIndex++
		pool.buff = append(pool.buff, d)
	}
}
*/
/////////////////////////////////////////////////////////////////////////////////////////////

type BuffDataPool struct {
	pools []*OneBuffDataPool
	lock  sync.Mutex
}

func (bp *BuffDataPool) Dump() string {
	str := ""
	for _, p := range bp.pools {
		if p != nil {
			str = fmt.Sprintf("%s%d:%s\n", str, p.cap, p.String())
		}
	}
	return str
}

func (bp *BuffDataPool) Init() {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	bp.pools = make([]*OneBuffDataPool, 33)
	for i := uint32(3); i < 33; i++ {
		pool := &OneBuffDataPool{}
		pool.cap = int(uint32(1<<i) - 1)
		pool.New = func() interface{} {
			data := MakeBuffData(pool.cap)
			data.pool = bp
			data.inUse = false
			return data
		}
		bp.pools[i] = pool
	}
}

func (bp *BuffDataPool) getPool(size int) *OneBuffDataPool {
	for i := 0; i < len(bp.pools); i++ {
		pool := bp.pools[i]
		if pool == nil {
			continue
		}
		if size <= pool.cap {
			return pool
		}
	}
	return nil
}

func (bp *BuffDataPool) Get(size int) *BuffData {
	if size >= math.MaxInt32 {
		return MakeBuffData(size)
	}

	bp.lock.Lock()
	pool := bp.getPool(size)
	bp.lock.Unlock()

	if pool != nil {
		data := pool.Get().(*BuffData)
		data.inUse = true
		data.index = 0
		data.size = size
		data.pool = bp
		return data
	}
	return nil
}

func (bp *BuffDataPool) Put(data *BuffData) {
	if data == nil || data.pool != bp {
		return
	}

	bp.lock.Lock()
	pool := bp.getPool(data.buffCap)
	if data.inUse == true {
		data.inUse = false
	} else {
		pool = nil
	}
	bp.lock.Unlock()

	if pool != nil {
		pool.Put(data)
	} else {
		//fmt.Printf("BuffDataPool 不能归还的数据 %v\n", data)
	}
}
