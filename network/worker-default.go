package network

import (
	"github.com/playnb/mustang/log"
	"context"
	"sync"
	"time"
)

type DefaultWorkerThread struct {
	isWorking       bool
	agent           IAgent
	timerAgent      ITimerAgent
	msgProcessor    IMsgProcessor
	allCallsRetChan map[string]chan interface{}

	worker *Worker
}

func (wt *DefaultWorkerThread) DoJob(f JobFunc, wait bool, ctx context.Context) error {
	return wt.worker.DoJob(f, wait, ctx)
}

func (wt *DefaultWorkerThread) Init(agent IAgent, msgProcessor IMsgProcessor, timerAgent ITimerAgent) {
	wt.agent = agent
	wt.timerAgent = timerAgent
	wt.msgProcessor = msgProcessor

	wt.allCallsRetChan = make(map[string]chan interface{})
	wt.isWorking = false
	wt.worker = &Worker{}
	wt.worker.Init()
}

func (wt *DefaultWorkerThread) Go(waitGroup *sync.WaitGroup) {
	if wt.isWorking {
		if waitGroup != nil {
			waitGroup.Done()
		}
		return
	}
	wt.isWorking = true
	go wt.run(waitGroup)
}

func (wt *DefaultWorkerThread) run(waitGroup *sync.WaitGroup) {

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
		case <-wt.agent.GetConnExitChan():
			//TODO: agentMsgChan,tcpMsgChan,没有关闭哦.....
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
		case data, ok := <-wt.agent.GetMsgChan():
			if ok {
				if wt.msgProcessor != nil {
					_, err := wt.msgProcessor.Handler(wt.agent, data, nil)
					if err != nil {
						log.Debug("%s msgProcessor error: %v", wt.agent, err)
					}
				}
			}
		}
	}
}
