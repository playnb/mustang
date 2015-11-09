package utils

import (
	"github.com/playnb/mustang/log"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

//全局唯一ID发生器

var generators map[uint64]*IdGenerator = make(map[uint64]*IdGenerator)

//获取snowflake的ID
func GenNextId(workerId, catalogId int64) (uint64, error) {
	if id, err := NewIdGenerator(workerId, catalogId); err == nil {
		return id.NextId()
	} else {
		return 0, err
	}
}

//获取snowflake的ID(不返回错误信息)
func GenEasyNextId(workerId, catalogId int64) uint64 {
	if id, err := NewIdGenerator(workerId, catalogId); err == nil {
		return id.EasyNextId()
	} else {
		return 0
	}
}

//构造一个新的id分配器
func NewIdGenerator(workerId, catalogId int64) (*IdGenerator, error) {
	if workerId > maxWorkerId || workerId < 0 {
		log.Error("workerId溢出,取值范围[0, %d]", maxWorkerId)
		return nil, errors.New(fmt.Sprintf("worker Id: %d error", workerId))
	}
	if catalogId > maxCatalogId || catalogId < 0 {
		log.Error("catalogId溢出,取值范围[0, %d]", maxCatalogId)
		return nil, errors.New(fmt.Sprintf("catalog Id: %d error", catalogId))
	}
	gid := makeGenID(workerId, catalogId)
	id, ok := generators[gid]
	if !ok {
		id = makeIdGenerator(workerId, catalogId)
		generators[gid] = id
	}
	return id, nil
}

const (
	twepoch       = int64(1420041600000) //时间参数基数
	workerIdBits  = uint(5)              //分配器ID占位
	catalogIdBits = uint(5)              //分类ID占位
	sequenceBits  = uint(12)             //毫秒内序列占位

	maxWorkerId  = -1 ^ (-1 << workerIdBits)
	maxCatalogId = -1 ^ (-1 << catalogIdBits)

	workerIdShift      = sequenceBits
	catalogIdShift     = sequenceBits + workerIdBits
	timestampLeftShift = sequenceBits + workerIdBits + catalogIdBits
	sequenceMask       = -1 ^ (-1 << sequenceBits)
	maxNextIdsNum      = 100
)

//ID构造器
type IdGenerator struct {
	sequence      int64
	lastTimestamp int64
	workerId      int64
	twepoch       int64
	catalogId     int64
	mutex         sync.Mutex
}

func makeGenID(workerId, catalogId int64) uint64 {
	return uint64((workerId << workerIdShift) | (catalogId << catalogIdShift))
}

func makeIdGenerator(workerId, catalogId int64) *IdGenerator {
	gen := &IdGenerator{}
	gen.workerId = workerId
	gen.catalogId = catalogId
	gen.lastTimestamp = -1
	gen.sequence = 0
	gen.twepoch = twepoch
	gen.mutex = sync.Mutex{}
	return gen
}

func timeGen() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func tilNextMillis(lastTimestamp int64) int64 {
	timestamp := timeGen()
	for timestamp <= lastTimestamp {
		time.Sleep(time.Millisecond)
		timestamp = timeGen()
	}
	return timestamp
}

func (id *IdGenerator) nextId() (uint64, error) {
	timestamp := timeGen()
	if timestamp < id.lastTimestamp {
		log.Error("IdGenerator分配ID错误，检测到时光倒流 %d 微秒", id.lastTimestamp-timestamp)
		return 0, errors.New(fmt.Sprintf("IdGenerator分配ID错误，检测到时光倒流 %d 微秒", id.lastTimestamp-timestamp))
	}
	if id.lastTimestamp == timestamp {
		id.sequence = (id.sequence + 1) & sequenceMask
		if id.sequence == 0 {
			timestamp = tilNextMillis(id.lastTimestamp)
		}
	} else {
		id.sequence = 0
	}
	id.lastTimestamp = timestamp
	return uint64(((timestamp - id.twepoch) << timestampLeftShift) | (id.catalogId << catalogIdShift) | (id.workerId << workerIdShift) | id.sequence), nil
}

func (id *IdGenerator) NextId() (uint64, error) {
	id.mutex.Lock()
	defer id.mutex.Unlock()
	return id.nextId()
}

func (id *IdGenerator) EasyNextId() uint64 {
	nid, _ := id.NextId()
	return nid
}

func (id *IdGenerator) NextIds(num int) ([]uint64, error) {
	if num > maxNextIdsNum || num < 0 {
		log.Error("批量生成ID太多了,范围[0, %d]", maxNextIdsNum)
		return nil, errors.New(fmt.Sprintf("NextIds num: %d error", num))
	}
	ids := make([]uint64, num)
	id.mutex.Lock()
	defer id.mutex.Unlock()
	var err error
	for i := 0; i < num; i++ {
		ids[i], err = id.nextId()
		if err != nil {
			return nil, errors.New(fmt.Sprintf("产生单个ID失败,%s,批量生成ID数%d", err.Error(), num))
		}
	}
	return ids, nil
}

// 获取本机的MAC地址
func (s *IdGenerator) mac() []byte {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error : " + err.Error())
	}
	for _, inter := range interfaces {
		mac := inter.HardwareAddr
		return mac
	}
	return nil
}

//获取主机名
func (s *IdGenerator) hostname() string {
	name, err := os.Hostname()
	if err != nil {
		log.Error("获取主机名失败: %v", err)
		return "default_host_"
	}
	return name
}

func TestSnow() {
	for i := 0; i < 10; i++ {
		log.Trace("Gen %d", GenEasyNextId(1, 1))
	}
}
