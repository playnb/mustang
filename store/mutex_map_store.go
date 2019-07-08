package store

import (
	"github.com/playnb/mustang/redis"
	"strconv"
	"sync"
)

func NewMutexMapStore(key string, newFunc NewICanStoreFunc) IStore {
	ins := &MutexMapStore{}
	ins.key = "Store:" + key
	ins.loaded = make(map[uint64]bool)
	ins.elements = make(map[uint64]ICanStore)
	ins.newFunc = newFunc
	return ins
}

type MutexMapStore struct {
	key      string
	loadAll  bool
	loaded   map[uint64]bool
	elements map[uint64]ICanStore
	newFunc  NewICanStoreFunc
	mutex    sync.Mutex
}

func (mgr *MutexMapStore) Load() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if mgr.loadAll {
		return
	}
	mgr.loadAll = true

	for k, v := range redis.GetPoolClient().HGetAll(mgr.key).Val() {
		uid, _ := strconv.ParseUint(k, 10, 64)
		if uid != 0 {
			elem := mgr.newFunc(uid)
			if elem != nil && elem.Deserialization([]byte(v)) {
				mgr.elements[elem.GetUniqueID()] = elem
			}
		}
	}
}

func (mgr *MutexMapStore) ForAll(f func(ICanStore)) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	for _, v := range mgr.elements {
		f(v)
	}
}

func (mgr *MutexMapStore) Get(uniqueID uint64) ICanStore {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if elem, ok := mgr.elements[uniqueID]; ok {
		if elem != nil {
			return elem.(ICanStore)
		}
		return nil
	}
	if !mgr.loadAll {
		if has, ok := mgr.loaded[uniqueID]; ok && has {
			return nil
		}
		mgr.loaded[uniqueID] = true
		buf, err := redis.GetPoolClient().HGet(mgr.key, strconv.FormatUint(uniqueID, 10)).Bytes()
		if err != nil {
			elem := mgr.newFunc(uniqueID)
			elem.Deserialization(buf)
			mgr.elements[uniqueID] = elem
			return elem
		}
	}
	return nil
}

func (mgr *MutexMapStore) Add(elem ICanStore) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	mgr.elements[elem.GetUniqueID()] = elem
}

func (mgr *MutexMapStore) Del(elem ICanStore) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	delete(mgr.elements, elem.GetUniqueID())
	redis.GetPoolClient().HDel(mgr.key, strconv.FormatUint(elem.GetUniqueID(), 10))
}

func (mgr *MutexMapStore) Save(elem ICanStore) {
	redis.GetPoolClient().HSet(mgr.key, strconv.FormatUint(elem.GetUniqueID(), 10), elem.Serializable())
}
