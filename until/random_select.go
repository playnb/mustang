package until

import (
	"fmt"
	"math/rand"
)

type RandomSelect struct {
	randomData map[interface{}]uint64
	maxWeight  uint64
}

func MakeRandomSelect() *RandomSelect {
	rs := &RandomSelect{}
	rs.randomData = make(map[interface{}]uint64)
	rs.maxWeight = 0

	return rs
}

func (rs *RandomSelect) Put(date interface{}, weight uint64) {
	rs.randomData[date] = weight
	rs.maxWeight += weight
}

func (rs *RandomSelect) Select() (date interface{}) {
	if rs.maxWeight == 0 {
		return nil
	}
	sl := rand.Uint64()%rs.maxWeight + 1
	a := uint64(0)
	b := uint64(0)
	for k, v := range rs.randomData {
		b += v
		if sl > a && sl <= b {
			return k
		}
		a += v
	}
	return nil
}

func (rs *RandomSelect) Pick() (date interface{}) {
	if rs.maxWeight == 0 {
		return nil
	}
	sl := rand.Uint64()%rs.maxWeight + 1
	a := uint64(0)
	b := uint64(0)
	for k, v := range rs.randomData {
		b += v
		if sl > a && sl <= b {
			delete(rs.randomData, k)
			return k
		}
		a += v
	}
	return nil
}

func (rs *RandomSelect) String() string {
	str := "["
	for k, v := range rs.randomData {
		str = str + fmt.Sprintf("(Prop:%d Data:%v)", v, k)
	}
	str = str + "]"
	return str
}
