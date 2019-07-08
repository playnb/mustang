package store

type ICanStore interface {
	GetUniqueID() uint64         //唯一标识
	Serializable() []byte        //序列化对象
	Deserialization([]byte) bool //反序列化对象
}

type IStore interface {
	Load()                         //加载全部数据
	ForAll(f func(ICanStore))      //遍历
	Get(uniqueID uint64) ICanStore //获取数据
	Add(elem ICanStore)            //添加数据
	Del(elem ICanStore)            //删除数据
	Save(elem ICanStore)           //数据存档
}

type NewICanStoreFunc func(uniqueID uint64) ICanStore

func NewDefaultStore(key string, newFunc NewICanStoreFunc) IStore {
	return NewSyncMapStore(key, newFunc)
}
