package mq

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type MyEntry struct {
	uniqueID uint64
}

func (e *MyEntry) UniqueID() uint64 {
	return e.uniqueID
}

func foo(topic string, entry MessageQueueEntry, data interface{}) {
	fmt.Printf("Topic:%s Entry:%d Data:%v \n", topic, entry.UniqueID(), data)
	time.Sleep(time.Microsecond)
}

func TestMessageQueue_AddEntry(t *testing.T) {
	m := New()

	add := func(n uint64) {
		for i := uint64(0); i < 1000; i++ {
			m.AddEntry(&MyEntry{uniqueID: i + n*10000})
			time.Sleep(time.Microsecond)
		}
	}
	del := func(n uint64) {
		for i := uint64(0); i < 1000; i++ {
			m.RemoveEntry(i + n*10000)
			time.Sleep(time.Microsecond)
		}
	}
	_ = del
	_ = add
	for i := uint64(0); i < 100; i++ {
		go add(i)
		go del(i)
	}
	time.Sleep(time.Second)

	sp := func() {
		for i := 0; i < 1000000; i++ {
			m.AddEntry(&MyEntry{uniqueID: 10})
			m.AddEntry(&MyEntry{uniqueID: 11})

			m.Subscribe(10, "T1", foo)
			m.Subscribe(11, "T1", foo)
			m.Subscribe(12, "T1", foo)
			m.Subscribe(11, "T2", foo)

			m.Publish("T1", "H111")
			m.Publish("T2", "H222")
			m.Publish("T3", "H333")

			m.Unsubscribe(11, "T2")
			m.Publish("T2", "M222")
			m.RemoveEntry(11)
			m.Publish("T1", "N222")
		}
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go sp()
	}
	wg.Done()

}
