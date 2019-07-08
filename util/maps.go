package util

import (
	"sync"
)

/*
没有模板,就简单封装一下类型,防止类型冲突

Store(key, value interface{})
LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
Load(key interface{}) (value interface{}, ok bool)
Delete(key interface{})
Range(f func(key, value interface{}) bool)
*/

//===================================================================================
//map[uint64]uint64
type MapUint64Uint64 struct {
	_map           sync.Map
	_length        uint64 //TODO: 并行访问的时候,长度并没有什么意义...
	_refresh       bool
	_value_add     uint64
	_value_refresh bool
}

func (m *MapUint64Uint64) Length() uint64 {
	if m._refresh {
		m._length = uint64(0)
		m._refresh = false
		m.Range(func(k uint64, v uint64) bool {
			m._length++
			return true
		})
	}
	return m._length
}
func (m *MapUint64Uint64) ValueAdd() uint64 {
	if m._value_refresh {
		m._value_add = uint64(0)
		m._value_refresh = false
		m.Range(func(k uint64, v uint64) bool {
			m._value_add = m._value_add + v
			return true
		})
	}
	return m._value_add
}
func (m *MapUint64Uint64) Store(key, value uint64) {
	m._map.Store(key, value)
	m._refresh = true
	m._value_refresh = true
}
func (m *MapUint64Uint64) LoadOrStore(key, value uint64) (uint64, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	m._refresh = true
	if actual != nil {
		return actual.(uint64), loaded
	}
	return 0, loaded
}
func (m *MapUint64Uint64) Load(key uint64) (uint64, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.(uint64), ok
	}
	return 0, ok
}
func (m *MapUint64Uint64) Delete(key uint64) {
	m._map.Delete(key)
	m._refresh = true
	m._value_refresh = true
}
func (m *MapUint64Uint64) Range(f func(key, value uint64) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(uint64), v.(uint64))
	})
}

//===================================================================================
//map[int]int
type MapIntInt struct {
	_map sync.Map
}

func (m *MapIntInt) Store(key, value int) {
	m._map.Store(key, value)
}
func (m *MapIntInt) LoadOrStore(key, value int) (int, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	if actual != nil {
		return actual.(int), loaded
	}
	return 0, loaded
}
func (m *MapIntInt) Load(key int) (int, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.(int), ok
	}
	return 0, ok
}
func (m *MapIntInt) Delete(key int) {
	m._map.Delete(key)
}
func (m *MapIntInt) Range(f func(key, value int) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(int), v.(int))
	})
}

//===================================================================================
// map[uint64]string
type MapUint64String struct {
	_map sync.Map
}

func (m *MapUint64String) Store(key uint64, value string) {
	m._map.Store(key, value)
}
func (m *MapUint64String) LoadOrStore(key uint64, value string) (string, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	if actual != nil {
		return actual.(string), loaded
	}
	return "", loaded
}
func (m *MapUint64String) Load(key uint64) (string, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.(string), ok
	}
	return "", ok
}
func (m *MapUint64String) Delete(key uint64) {
	m._map.Delete(key)
}
func (m *MapUint64String) Range(f func(key uint64, value string) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(uint64), v.(string))
	})
}

//===================================================================================
//map[string]uint64
type MapStringUint64 struct {
	_map sync.Map
}

func (m *MapStringUint64) Store(key string, value uint64) {
	m._map.Store(key, value)
}

func (m *MapStringUint64) LoadOrStore(key string, value uint64) (uint64, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	if actual != nil {
		return actual.(uint64), loaded
	}
	return 0, loaded
}

func (m *MapStringUint64) Load(key string) (uint64, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.(uint64), ok
	}
	return 0, ok
}

func (m *MapStringUint64) Delete(key string) {
	m._map.Delete(key)
}

func (m *MapStringUint64) Range(f func(key string, value uint64) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(uint64))
	})
}
/*
//===================================================================================
// map[uint64]*msg.T_ExpandBaseUserData
type MapUint64ExpandBaseUserData struct {
	_map sync.Map
}

func (m *MapUint64ExpandBaseUserData) Store(key uint64, value *msg.T_ExpandBaseUserData) {
	m._map.Store(key, value)
}
func (m *MapUint64ExpandBaseUserData) LoadOrStore(key uint64, value *msg.T_ExpandBaseUserData) (*msg.T_ExpandBaseUserData, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	if actual != nil {
		return actual.(*msg.T_ExpandBaseUserData), loaded
	}
	return nil, loaded
}
func (m *MapUint64ExpandBaseUserData) Load(key uint64) (*msg.T_ExpandBaseUserData, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.(*msg.T_ExpandBaseUserData), ok
	}
	return nil, ok
}
func (m *MapUint64ExpandBaseUserData) Delete(key uint64) {
	m._map.Delete(key)
}
func (m *MapUint64ExpandBaseUserData) Range(f func(key uint64, value *msg.T_ExpandBaseUserData) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(uint64), v.(*msg.T_ExpandBaseUserData))
	})
}

//===================================================================================
// map[uint64]*msg.RoomVideo

//===================================================================================
// map[uint64]*msg.T_PhoneUser
type MapUint64PhoneUser struct {
	_map sync.Map
}

func (m *MapUint64PhoneUser) Store(key uint64, value *msg.T_PhoneUser) {
	m._map.Store(key, value)
}
func (m *MapUint64PhoneUser) LoadOrStore(key uint64, value *msg.T_PhoneUser) (*msg.T_PhoneUser, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	if actual != nil {
		return actual.(*msg.T_PhoneUser), loaded
	}
	return nil, loaded
}
func (m *MapUint64PhoneUser) Load(key uint64) (*msg.T_PhoneUser, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.(*msg.T_PhoneUser), ok
	}
	return nil, ok
}
func (m *MapUint64PhoneUser) Delete(key uint64) {
	m._map.Delete(key)
}
func (m *MapUint64PhoneUser) Range(f func(key uint64, value *msg.T_PhoneUser) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(uint64), v.(*msg.T_PhoneUser))
	})
}
*/
//===================================================================================
// map[uint32]util.RandomSelect
type MapUint32RandomSelect struct {
	_map sync.Map
}

func (m *MapUint32RandomSelect) Store(key uint32, value RandomSelect) {
	m._map.Store(key, value)
}
func (m *MapUint32RandomSelect) LoadOrStore(key uint32, value RandomSelect) (RandomSelect, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	if actual != nil {
		return actual.(RandomSelect), loaded
	}
	return nil, loaded
}
func (m *MapUint32RandomSelect) Load(key uint32) (RandomSelect, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.(RandomSelect), ok
	}
	return nil, ok
}
func (m *MapUint32RandomSelect) Delete(key uint32) {
	m._map.Delete(key)
}
func (m *MapUint32RandomSelect) Range(f func(key uint32, value RandomSelect) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(uint32), v.(RandomSelect))
	})
}

//===================================================================================
//map[uint64][]uint64
type MapUint64SliceUint64 struct {
	_map sync.Map
}

func (m *MapUint64SliceUint64) Store(key uint64, value []uint64) {
	m._map.Store(key, value)
}
func (m *MapUint64SliceUint64) LoadOrStore(key uint64, value []uint64) ([]uint64, bool) {
	actual, loaded := m._map.LoadOrStore(key, value)
	if actual != nil {
		return actual.([]uint64), loaded
	}
	return nil, loaded
}
func (m *MapUint64SliceUint64) Load(key uint64) ([]uint64, bool) {
	value, ok := m._map.Load(key)
	if value != nil {
		return value.([]uint64), ok
	}
	return nil, ok
}
func (m *MapUint64SliceUint64) Delete(key uint64) {
	m._map.Delete(key)
}
func (m *MapUint64SliceUint64) Range(f func(key uint64, value []uint64) bool) {
	m._map.Range(func(k, v interface{}) bool {
		return f(k.(uint64), v.([]uint64))
	})
}

//===================================================================================
