package store

import (
	"github.com/playnb/mustang/redis"
	"strconv"
	"sync"
)

func NewSyncMapStore(key string, newFunc NewICanStoreFunc) IStore {
	ins := &SyncMapStore{}
	ins.key = "Store:" + key
	ins.elements = &sync.Map{}
	ins.newFunc = newFunc
	return ins
}

type SyncMapStore struct {
	key      string
	loaded   bool
	elements *sync.Map
	newFunc  NewICanStoreFunc
}

func (mgr *SyncMapStore) Load() {
	if mgr.loaded {
		return
	}
	mgr.loaded = true

	for k, v := range redis.GetPoolClient().HGetAll(mgr.key).Val() {
		uid, _ := strconv.ParseUint(k, 10, 64)
		if uid != 0 {
			elem := mgr.newFunc(uid)
			if elem != nil && elem.Deserialization([]byte(v)) {
				mgr.elements.Store(elem.GetUniqueID(), elem)
			}
		}
	}
}

func (mgr *SyncMapStore) ForAll(f func(ICanStore)) {
	mgr.elements.Range(func(key, value interface{}) bool {
		if value != nil {
			f(value.(ICanStore))
		}
		return true
	})
}

func (mgr *SyncMapStore) Get(uniqueID uint64) ICanStore {
	if elem, ok := mgr.elements.Load(uniqueID); ok {
		if elem != nil {
			return elem.(ICanStore)
		}
		return nil
	}
	if !mgr.loaded {
		buf, err := redis.GetPoolClient().HGet(mgr.key, strconv.FormatUint(uniqueID, 10)).Bytes()
		if err != nil {
			elem := mgr.newFunc(uniqueID)
			elem.Deserialization(buf)
			mgr.elements.Store(uniqueID, elem)
			return elem
		}
	}
	return nil
}

func (mgr *SyncMapStore) Add(elem ICanStore) {
	mgr.elements.Store(elem.GetUniqueID(), elem)
}

func (mgr *SyncMapStore) Del(elem ICanStore) {
	mgr.elements.Delete(elem.GetUniqueID())
	redis.GetPoolClient().HDel(mgr.key, strconv.FormatUint(elem.GetUniqueID(), 10))
}

func (mgr *SyncMapStore) Save(elem ICanStore) {
	redis.GetPoolClient().HSet(mgr.key, strconv.FormatUint(elem.GetUniqueID(), 10), elem.Serializable())
}
