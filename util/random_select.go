package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type RandomSelect interface {
	Put(date interface{}, weight uint64)
	Select() interface{} //随机选择一个
	Pick() interface{}   //随机选择一个并删除
}

type randomSelect struct {
	randomData map[interface{}]uint64
	maxWeight  uint64
}

func MakeRandomSelect() RandomSelect {
	rs := &randomSelect{}
	rs.randomData = make(map[interface{}]uint64)
	rs.maxWeight = 0

	return rs
}

func (rs *randomSelect) Put(date interface{}, weight uint64) {
	if w, ok := rs.randomData[date]; ok == true {
		rs.maxWeight -= w
	}
	rs.randomData[date] = weight
	rs.maxWeight += weight
}

func (rs *randomSelect) Select() (date interface{}) {
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

func (rs *randomSelect) Pick() (date interface{}) {
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
			rs.maxWeight = rs.maxWeight - v
			return k
		}
		a += v
	}
	return nil
}

func (rs *randomSelect) String() string {
	str := "["
	for k, v := range rs.randomData {
		str = str + fmt.Sprintf("(Prop:%d Data:%v)", v, k)
	}
	str = str + "]"
	return str
}

type IDPropRandomSelect struct {
	RandomSelect
}

func (rs *IDPropRandomSelect) Init() {
	rs.RandomSelect = MakeRandomSelect()
}

func (rs *IDPropRandomSelect) Parse(str string, s1 string, s2 string) {
	ss1 := strings.Split(str, s2)
	for _, str1 := range ss1 {
		ss2 := strings.Split(str1, s1)
		if len(ss2) == 2 {
			id, _ := strconv.ParseUint(ss2[0], 10, 64)
			prop, _ := strconv.ParseUint(ss2[1], 10, 64)
			rs.Put(id, prop)
		}
	}
}
