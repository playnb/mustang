package network

import (
	"cell/common/mustang/log"
	"context"
	"errors"
	"sync"
)

type JobFunc func()
type Job struct {
	f      JobFunc
	end    chan int
	cancel bool
}

var JobTimeOut = errors.New("JobTimeOut")

type WorkerThread interface {
	Go(waitGroup *sync.WaitGroup)
	DoJob(f JobFunc, wait bool, ctx context.Context) error
}

type Worker struct {
	JobChan         chan *Job
	allCallsRetChan map[string]chan interface{}
	workingMutex    sync.Mutex
}

func (wt *Worker) Init() {
	wt.JobChan = make(chan *Job, 1000)
	wt.allCallsRetChan = make(map[string]chan interface{})
}

func (wt *Worker) DoJob(f JobFunc, wait bool, ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	job := &Job{
		f:      f,
		cancel: false,
	}
	if wait {
		job.end = make(chan int)
	}

	select {
	case wt.JobChan <- job:
		{
			if job.end != nil {
				select {
				case <-job.end:
					return nil
				case <-ctx.Done():
					return JobTimeOut
				}
			}
			return nil
		}
	case <-ctx.Done():
		{
			return JobTimeOut
		}
	}
}

func (wt *Worker) RunJob(j *Job) {
	if j.cancel {
	} else {
		func() {
			defer log.PrintPanicStack()
			j.f()
		}()
	}
	if j.end != nil {
		close(j.end)
	}
}

func (wt *Worker) FinishAllJob() {
	for _, c := range wt.allCallsRetChan {
		close(c)
		//return true
	}
}
