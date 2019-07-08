package event

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var _dispatcher *Dispatcher
var Event1 = "TestEvent1"
var Event2 = "TestEvent2"

type EventData struct {
	index []int
}

func init() {
	_dispatcher = NewDispatcher()
}

func TestDispatcher(t *testing.T) {
	tf := func(priority int64, index int) {
		_dispatcher.AddAction(Event1, func(i interface{}) {
			d := i.(*EventData)
			d.index = append(d.index, index)
			//t.Logf("C:%d priority: %d  index:%d\n", len(d.index), priority, index)
		}, priority)
	}
	compare := func(a []int, b []int) bool {
		t.Logf("TestA: %v", a)
		t.Logf("TestB: %v", b)
		if len(a) != len(b) {
			return false
		}
		for k, v := range a {
			if v != b[k] {
				return false
			}
		}
		return true
	}

	tf(0, 1)
	tf(1, 2)
	tf(0, 3)
	tf(-1, 4)
	tf(10, 5)
	tf(-1, 6)
	tf(-1, 7)
	tf(0, 8)
	tf(-2, 9)
	tf(1, 10)

	tt := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		data := &EventData{}
		if rand.Intn(10) > 5 {
			_dispatcher.Trigger(Event1, data)
			if compare(data.index, []int{1, 3, 8, 2, 10, 5, 9, 4, 6, 7}) == false {
				t.Fail()
			}
		} else {
			_dispatcher.Trigger(Event2, data)
			if compare(data.index, []int{}) == false {
				t.Fail()
			}
		}
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go tt(wg)
	}

	wg.Wait()

}

func TestDispatcher2(t *testing.T) {
	c1 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		for i := 0; i < 1000; i++ {
			_dispatcher.AddAction(fmt.Sprintf("%s_%d", Event2, rand.Int63n(1000)), func(i interface{}) {
			}, rand.Int63n(100)-50)
		}
		t.Log("c1")
	}
	c2 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		for i := 0; i < 1000; i++ {
			_dispatcher.Trigger(fmt.Sprintf("%s_%d", Event2, rand.Int63n(1000)), struct{}{})
		}
		t.Log("c2")
	}

	wg := &sync.WaitGroup{}
	t.Log("Begin")
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		if rand.Int63n(10) > 5 {
			go c1(wg)
		} else {
			go c2(wg)
		}
	}
	t.Log("Wait")
	wg.Wait()
}

func TestDispatcher3(t *testing.T) {
	pass := false
	_dispatcher.AddAction(Event1, func(i interface{}) {
		_dispatcher.Trigger(Event2, nil)
	}, 0)
	_dispatcher.AddAction(Event2, func(i interface{}) {
		//_dispatcher.Trigger(Event2, nil)
		t.Log("Event2的处理函数")
		pass = true
	}, 0)
	_dispatcher.Trigger(Event1, nil)
	if !pass {
		t.Fail()
	}
}
