package cell_base

import (
	"sync"
)

type RpcClientMap struct {
	_map sync.Map
}

func (m *RpcClientMap) Store(uid uint64, client *RpcClient) {
	m._map.Store(uid, client)
}

func (m *RpcClientMap) Load(id uint64) (*RpcClient, bool) {
	client, ok := m._map.Load(id)
	if ok {
		return client.(*RpcClient), true
	}
	return nil, false
}

func (m *RpcClientMap) Delete(id uint64) {
	m._map.Delete(id)
}

func (m *RpcClientMap) Range(foo func(*RpcClient) bool) {
	m._map.Range(func(key, value interface{}) bool {
		return foo(value.(*RpcClient))
	})
}

