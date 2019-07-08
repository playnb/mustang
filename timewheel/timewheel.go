package timewheel

//借鉴https://github.com/ouqiang/timewheel

import (
	"github.com/playnb/mustang/util"
	"container/list"
	_ "fmt"
	"time"
)

type Job func(TaskData)
type TaskKey interface{}
type TaskData interface{}

type _Task struct {
	key      TaskKey
	active   bool //是否激活状态
	pos      int
	delay    time.Duration // 延迟时间
	circle   int           // 时间轮需要转动几圈
	data     TaskData
	sequence uint64
}

type _OpTask struct {
	add  bool
	task *_Task
	key  TaskKey
}

type ITimeWheel interface {
	Start()
	Stop()
	Join()
	AddTimerDelaySecond(delaySceond time.Duration, key TaskKey, data TaskData)
	AddTimerTimestamp(timestamp int64, key TaskKey, data TaskData)
	RemoveTimer(key TaskKey)
	RefreshAll()
}

type _TimeWheel struct {
	interval time.Duration // 指针每隔多久往前移动一格
	slotNum  int           // 槽数量
	job      Job           // 定时器回调函数
	//addTaskChannel    chan *Task    // 新增任务channel
	//removeTaskChannel chan TaskKey  // 删除任务channel
	operationChannel chan *_OpTask // 新增/删除任务channel
	stopChannel      chan bool     // 停止定时器channel
	joinChannel      chan bool     // Join定时器channel
	refreshAll       chan bool
	tasks            map[TaskKey]map[uint64]*_Task
	sequence         uint64
	startTime        int64
	runTime          int64

	slots      []*list.List // 时间轮槽
	ticker     *time.Ticker
	currentPos int
}

//New创建一个time wheel
//interval: 运行最小间隔(秒)
//slotNum: 时间轮槽数(越大，产生的cycle越小，消耗内存越大)
//job: 定时的回调函数
func New(interval time.Duration, slotNum int, job Job) ITimeWheel {
	tw := &_TimeWheel{
		interval: interval * time.Second,
		slotNum:  slotNum,
		job:      job,
		//addTaskChannel:    make(chan *Task, 1024),
		//removeTaskChannel: make(chan TaskKey, 1024),
		operationChannel: make(chan *_OpTask, 1024),
		stopChannel:      make(chan bool),
		tasks:            make(map[TaskKey]map[uint64]*_Task),
	}

	tw.currentPos = 0
	tw.slots = make([]*list.List, tw.slotNum)
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
	return tw
}

func (tw *_TimeWheel) RefreshAll() {
	tw.refreshAll <- true
}

//Start 启动时间轮
func (tw *_TimeWheel) Start() {
	tw.startTime = time.Now().UnixNano()
	tw.ticker = time.NewTicker(tw.interval)
	tw.joinChannel = make(chan bool)
	go func() {
		for {
			select {
			case <-tw.ticker.C:
				tw.tickHandler()
				/*
				case task := <-tw.addTaskChannel:
					tw.addTask(task)
				case key := <-tw.removeTaskChannel:
					tw.removeTask(key)
				*/
			case op := <-tw.operationChannel:
				if op.add {
					tw.addTask(op.task)
				} else {
					tw.removeTask(op.key)
				}
			case <-tw.stopChannel:
				tw.ticker.Stop()
				close(tw.joinChannel)
				return
			case <-tw.refreshAll:
				for i := 0; i < tw.slotNum; i++ {
					tw.tickHandler()
				}
			}
		}
	}()
}

//Stop 停止时间轮
func (tw *_TimeWheel) Stop() {
	close(tw.stopChannel)
}

//Join 等待时间轮结束
func (tw *_TimeWheel) Join() {
	if tw.joinChannel != nil {
		<-tw.joinChannel
	}
}

//AddTimerDelaySecond 按照延迟(秒)增加定时
//delaySceond: 定时延迟(秒)
//key: 定时标识(删除用)
//data: 定时数据
func (tw *_TimeWheel) AddTimerDelaySecond(delaySceond time.Duration, key TaskKey, data TaskData) {
	if delaySceond <= 0 || key == nil {
		return
	}
	delaySceond = delaySceond + 1 //由于TimeWheel在Start的时候并不是整秒，可能导致AddTimer的时候虽然物理时间已经是T+1秒，但TimeWheel认为的实际时间在T秒
	op := &_OpTask{
		add:  true,
		task: &_Task{delay: delaySceond * time.Second, key: key, data: data},
	}
	tw.operationChannel <- op
}

//AddTimerDelaySecond 按照时间戳增加定时
//timestamp: 时间戳
//key: 定时标识(删除用)
//data: 定时数据
func (tw *_TimeWheel) AddTimerTimestamp(timestamp int64, key TaskKey, data TaskData) {
	delaySceond := time.Duration(timestamp - tw.runTime)
	tw.AddTimerDelaySecond(delaySceond, key, data)
}

//RemoveTimer 删除定时器
//key: 定时标识
func (tw *_TimeWheel) RemoveTimer(key TaskKey) {
	if key == nil {
		return
	}
	op := &_OpTask{
		add: false,
		key: key,
	}
	tw.operationChannel <- op
}

func (tw *_TimeWheel) tickHandler() {
	tw.runTime = util.NowTimestamp()
	//fmt.Printf("Now:%v\n", time.Now().Unix())
	l := tw.slots[tw.currentPos%tw.slotNum]
	tw.scanAndRunTask(l)
	tw.currentPos++
}

func (tw *_TimeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*_Task)
		if task.active {
			if task.circle > 0 {
				task.circle--
				e = e.Next()
				continue
			}

			go tw.job(task.data)
		}
		next := e.Next()
		l.Remove(e)
		//delete(tw.tasks, task.key)
		{
			if m, ok := tw.tasks[task.key]; ok {
				delete(m, task.sequence)
				if len(m) == 0 {
					delete(tw.tasks, task.key)
				}
			}
		}
		e = next
	}
}

func (tw *_TimeWheel) addTask(task *_Task) {
	tw.sequence++

	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle
	task.pos = pos
	task.active = true
	task.sequence = tw.sequence

	tw.slots[pos].PushBack(task)
	if _, ok := tw.tasks[task.key]; ok == false {
		tw.tasks[task.key] = make(map[uint64]*_Task)
	}
	tw.tasks[task.key][task.sequence] = task
}

func (tw *_TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(tw.interval.Seconds())
	circle = int(delaySeconds / intervalSeconds / tw.slotNum)
	pos = int(tw.currentPos+delaySeconds/intervalSeconds) % tw.slotNum
	return
}

func (tw *_TimeWheel) removeTask(key TaskKey) {
	if task, ok := tw.tasks[key]; ok {
		//task.active = false
		for _, v := range task {
			v.active = false
		}
	}
}
