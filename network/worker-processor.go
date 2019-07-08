package network

import (
	"cell/common/mustang/log"
	"cell/common/mustang/util"
	"context"
	"sync"
	"time"
)

type ProcessorWorkerThread struct {
	Name      string
	isWorking bool
	exitChan  chan bool
	dataChan  chan *util.BuffData

	netAgent     SampleAgent
	Agent        interface{}   //消息处理主体
	msgProcessor IMsgProcessor //消息处理框架
	timerAgent   ITimerAgent   //Timer时间对象

	worker *Worker
}

func (wt *ProcessorWorkerThread) Init(netAgent SampleAgent, msgProcessor IMsgProcessor, Agent interface{}) {
	wt.timerAgent, _ = Agent.(ITimerAgent)
	wt.msgProcessor = msgProcessor
	wt.Agent = Agent
	wt.netAgent = netAgent

	wt.isWorking = false

	wt.worker = &Worker{}
	wt.worker.Init()
	wt.exitChan = make(chan bool)
	wt.dataChan = make(chan *util.BuffData, 1000)
}
func (wt *ProcessorWorkerThread) String() string {
	return "ProcessorWorkerThread[" + wt.Name + "]"
}
func (wt *ProcessorWorkerThread) DoJob(f JobFunc, wait bool, ctx context.Context) error {
	return wt.worker.DoJob(f, wait, ctx)
}
func (wt *ProcessorWorkerThread) Go(waitGroup *sync.WaitGroup) {
	if wt.isWorking {
		if waitGroup != nil {
			waitGroup.Done()
		}
		return
	}
	wt.isWorking = true
	go wt.run(waitGroup)
}
func (wt *ProcessorWorkerThread) PutData(data *util.BuffData) {
	wt.dataChan <- data
}
func (wt *ProcessorWorkerThread) Terminate() {
	close(wt.exitChan)
}

func (wt *ProcessorWorkerThread) run(waitGroup *sync.WaitGroup) {
	defer log.PrintPanicStack()
	defer wt.worker.FinishAllJob()
	defer func() {
		if waitGroup != nil {
			waitGroup.Done()
		}
		wt.isWorking = false
	}()

	ts := time.NewTicker(time.Second)
	tm := time.NewTicker(time.Minute)
	tfm := time.NewTicker(time.Minute * 5)
	th := time.NewTicker(time.Hour)
	defer ts.Stop()
	defer tm.Stop()
	defer tfm.Stop()
	defer th.Stop()

	for {
		select {
		case <-wt.exitChan:
			return
		case <-ts.C:
			if wt.timerAgent != nil {
				wt.timerAgent.SecondTimerFunc()
			}
		case <-tm.C:
			if wt.timerAgent != nil {
				wt.timerAgent.MinuteTimerFunc()
			}
		case <-tfm.C:
			if wt.timerAgent != nil {
				wt.timerAgent.FiveMinuteTimerFunc()
			}
		case <-th.C:
			if wt.timerAgent != nil {
				wt.timerAgent.HourTimerFunc()
			}
		case j := <-wt.worker.JobChan:
			wt.worker.RunJob(j)
		case data, ok := <-wt.dataChan:
			if ok {
				if wt.msgProcessor != nil {
					_, err := wt.msgProcessor.Handler(wt.netAgent, data, wt.Agent)
					if err != nil {
						log.Debug("%s msgProcessor error: %v", wt.String(), err)
					}
				}
			}
		}
	}
}
