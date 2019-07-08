// package event

/*
简单实现一个事件派发机制
 */
package event

import (
	"math"
	"sort"
	"sync"
)

var defaultDispatcher *Dispatcher = NewDispatcher()

func DefaultDispatcher() *Dispatcher {
	return defaultDispatcher
}

//====================================================================================

type Action func(interface{})

type actionElement struct {
	action   Action
	priority int64
	index    int
}

//====================================================================================

func NewActionList() *ActionList {
	al := &ActionList{}
	al.alwaysAdd = 0
	return al
}

type ActionList struct {
	actions   []*actionElement
	alwaysAdd int
	mutex     sync.RWMutex
}

//Add 增加处理函数
func (al *ActionList) Add(action Action, priority int64) int {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	if priority < 0 {
		priority = math.MaxInt64 + priority
	}
	al.alwaysAdd++
	al.actions = append(al.actions, &actionElement{action: action, priority: priority, index: al.alwaysAdd})
	sort.Slice(al.actions, func(i, j int) bool {
		return al.actions[i].priority < al.actions[j].priority
	})
	return al.alwaysAdd
}

func (al *ActionList) RemoveAll() {
	al.actions = nil
}

func (al *ActionList) Invoke(arg interface{}) {
	if len(al.actions) == 0 {
		return
	}

	ff := make([]*actionElement, 0, len(al.actions))
	al.mutex.RLock()
	defer al.mutex.RUnlock()

	for _, f := range al.actions {
		ff = append(ff, f)
	}

	for _, f := range ff {
		if f != nil {
			f.action(arg)
		}
	}
}

//====================================================================================

func NewDispatcher() *Dispatcher {
	var ins = &Dispatcher{}
	ins.events = make(map[string]*ActionList)
	return ins
}

type Dispatcher struct {
	events map[string]*ActionList
	mutex  sync.RWMutex
}

//AddAction 注册事件处理函数
//priority:优先级(定义事件处理函数调用的顺序，同数字，先添加先执行，如果是正数，数字越小处理函数越先执行，如果是负数，数字越小表示越晚执行)
func (d *Dispatcher) AddAction(event string, action Action, priority int64) int {
	d.mutex.Lock()
	evt, _ := d.events[event]
	if evt == nil {
		evt = NewActionList()
		d.events[event] = evt
		d.mutex.Unlock()
	} else {
		d.mutex.Unlock()
	}
	return evt.Add(action, priority)
}

//Trigger 触发一个事件
//source表示事件的数据(作为参数传到事件处理函数中)
func (d *Dispatcher) Trigger(event string, source interface{}) {
	d.mutex.RLock()
	evt, _ := d.events[event]
	d.mutex.RUnlock()
	if evt != nil {
		evt.Invoke(source)
	}
}
