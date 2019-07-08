package mq

import "sync"

type MessageQueueEntry interface {
	UniqueID() uint64
}

type MessageQueueAction func(topic string, entry MessageQueueEntry, data interface{})

func New() *MessageQueue {
	return &MessageQueue{}
}

type messageQueueData struct {
	entry       MessageQueueEntry
	interesting []string
}

type MessageQueue struct {
	allEntries sync.Map
	allTopics  sync.Map

	lock sync.Mutex
}

func (mq *MessageQueue) Init() {

}

//增加一个订阅对象
func (mq *MessageQueue) AddEntry(entry MessageQueueEntry) {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.doRemoveEntry(entry.UniqueID())
	mq.doAddEntry(entry)
}

//删除一个订阅对象
func (mq *MessageQueue) RemoveEntry(id uint64) {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.doRemoveEntry(id)
}

//订阅主题
func (mq *MessageQueue) Unsubscribe(id uint64, topic string) {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.doUnsubscribe(id, topic)
}

//取消对象的所有filter返回true的订阅
func (mq *MessageQueue) UnsubscribeAll(id uint64, filter func(string) bool) {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.doUnsubscribeAll(id, filter)
}

//取消订阅主题
func (mq *MessageQueue) Subscribe(id uint64, topic string, action MessageQueueAction) {
	mq.lock.Lock()
	defer mq.lock.Unlock()
	mq.doSubscribe(id, topic, action)
}

//发布主题
func (mq *MessageQueue) Publish(topic string, data interface{}) {
	mq.doPublish(topic, data)
}

func (mq *MessageQueue) doAddEntry(entry MessageQueueEntry) {
	mq.allEntries.Store(entry.UniqueID(), &messageQueueData{
		entry: entry,
	})
}

func (mq *MessageQueue) doRemoveEntry(id uint64) {
	e := mq.findEntry(id)
	if e != nil {
		for _, v := range e.interesting {
			mq.doUnsubscribe(id, v)
		}
		e.interesting = nil
		mq.allEntries.Delete(id)
	}
}

func (mq *MessageQueue) findEntry(id uint64) *messageQueueData {
	e, ok := mq.allEntries.Load(id)
	if ok {
		return e.(*messageQueueData)
	}
	return nil
}

func (mq *MessageQueue) doUnsubscribe(id uint64, topic string) {
	tm, ok := mq.allTopics.Load(topic)
	if ok {
		tm.(*sync.Map).Delete(id)
	}
}

func (mq *MessageQueue) doUnsubscribeAll(id uint64, filter func(string) bool) {
	e := mq.findEntry(id)
	if e != nil {
		var f []string
		for _, v := range e.interesting {
			if filter(v) {
				mq.doUnsubscribe(id, v)
			} else {
				f = append(f, v)
			}
		}
		e.interesting = f
	}
}

func (mq *MessageQueue) doSubscribe(id uint64, topic string, action MessageQueueAction) {
	e := mq.findEntry(id)
	if e == nil {
		return
	}
	tm, ok := mq.allTopics.LoadOrStore(topic, &sync.Map{})
	if ok {
		//以前就有这个topic
	}
	tm.(*sync.Map).Store(id, action)
	e.interesting = append(e.interesting, topic)
}

func (mq *MessageQueue) doPublish(topic string, data interface{}) {
	tm, ok := mq.allTopics.Load(topic)
	if ok {
		tm.(*sync.Map).Range(func(key, value interface{}) bool {
			id := key.(uint64)
			fn := value.(MessageQueueAction)
			e := mq.findEntry(id)
			if e != nil {
				fn(topic, e.entry, data)
			}
			return true
		})
	}
}
