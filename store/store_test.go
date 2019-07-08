package store

import (
	"github.com/playnb/mustang/redis"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/proto/testdata"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type _StoreElem struct {
	_uniqueID uint64
	_data     *testdata.GoTest
}

func (elem *_StoreElem) GetUniqueID() uint64 { //唯一标识
	return elem._uniqueID
}
func (elem *_StoreElem) Serializable() []byte { //序列化对象
	buf, err := proto.Marshal(elem._data)
	if err != nil {
		return nil
	}
	return buf
}
func (elem *_StoreElem) Deserialization(buf []byte) bool { //反序列化对象
	elem._data = &testdata.GoTest{}
	proto.Unmarshal(buf, elem._data)
	return true
}

func _NewStoreElem(uid uint64) ICanStore {
	elem := &_StoreElem{
		_uniqueID: uid,
		_data:     &testdata.GoTest{},
	}
	return elem
}

func init() {
	redis.InitRedisPool("127.0.0.1:6379", "")

	go func() {}()
}

func TestMutexMapStore(t *testing.T) {
	RunMutes(NewMutexMapStore("Test:Store", _NewStoreElem), t)
}

func TestSyncMapStore(t *testing.T) {
	RunMutes(NewSyncMapStore("Test:Store", _NewStoreElem), t)
}

func RunMutes(store IStore, t *testing.T) {
	A := int32(math.MaxInt32)
	B := int32(0)
	e1 := _NewStoreElem(1).(*_StoreElem)
	e2 := _NewStoreElem(2).(*_StoreElem)
	var d1 *testdata.GoTest
	e1._data.Param = proto.Int32(A)
	e2._data.Param = proto.Int32(B)
	var wg sync.WaitGroup
	terminate := false

	wg.Add(1)
	go func() {
		for !terminate {
			d1 = e1._data
			d1 = e2._data
		}
		wg.Done()
	}()
	time.Sleep(time.Millisecond)

	c1 := 0
	c2 := 0
	wg.Add(1)
	go func() {
		for i := 0; i < 1000000000; i++ {
			if d1.GetParam() == A {
				c1 += 1
				continue
			}
			if d1.GetParam() == B {
				c2 += 1
				continue
			}
			t.Log("Fail")
			t.Fail()
			break
		}
		terminate = true
		wg.Done()
	}()

	wg.Wait()
	t.Log(fmt.Sprintf("c1:%d c2:%d totla:%d\n", c1, c2, c1+c2))

	/*
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(e1._data)), unsafe.Pointer(e2._data))
	data := (*testdata.GoTest)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(e1._data))))
	t.Log(data.GetParam())
	*/
}

func BenchmarkAtomicStore(b *testing.B) {
	e1 := _NewStoreElem(1).(*_StoreElem)
	e2 := _NewStoreElem(2).(*_StoreElem)
	e2._data.Param = proto.Int32(1991)

	for i := 0; i < b.N; i++ {
		//atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(e1._data)), unsafe.Pointer(e2._data))
		(*testdata.GoTest)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(e1._data)))).GetParam()
	}
}

func BenchmarkDirectStore(b *testing.B) {
	e1 := _NewStoreElem(1).(*_StoreElem)
	e2 := _NewStoreElem(2).(*_StoreElem)
	e2._data.Param = proto.Int32(1991)
	for i := 0; i < b.N; i++ {
		e1._data = e2._data
		e1._data.GetParam()
	}
}
