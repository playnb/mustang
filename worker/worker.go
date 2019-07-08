package worker

import (
	"github.com/playnb/mustang/log"
	"fmt"
)

func DefaultPool() *Pool {
	return pool
}

var pool *Pool

func init() {
	pool = &Pool{}
}

type JobFunc func()

//Job
type Job struct {
	f   JobFunc
	end chan int
}

//Pool
type Pool struct {
	workerCount int
	running     bool
	workers     []*Worker
	maxJob      int
}

func (p *Pool) Start(count int, recoverFunc func()) {
	if p.running || count == 0 {
		return
	}
	p.workerCount = count
	p.maxJob = 1000
	p.workers = make([]*Worker, p.workerCount, p.workerCount)
	for i := 0; i < p.workerCount; i++ {
		w := &Worker{}
		w.id = i
		w.job = make(chan *Job, p.maxJob)
		w.recoverFunc = recoverFunc
		p.workers[w.id] = w
		go w.work()
	}
	p.running = true
}

//通过Worker执行JobFunc
//id: JobFunc的归属，相同的id会在同一个gorountine执行
//wait: 是否阻塞，直到执行结束
//f: 执行的Job
func (p *Pool) Do(id int, wait bool, f JobFunc) {
	if p.running == false {
		return
	}
	rawID := id

	j := &Job{}
	j.f = f
	j.end = make(chan int)
	id = id % p.workerCount

	if len(p.workers[id].job) > p.maxJob/2 {
		log.Error("[Worker] %d(%d)工作负载过大 %d(%d)(%v) ", id, rawID, len(p.workers[id].job), p.maxJob, p.workers[id].working)
	}
	p.workers[id].job <- j

	if wait {
		<-j.end
	}
}

func (p *Pool) Dump() {
	for _, v := range p.workers {
		fmt.Printf("[Dump][Worker:%d][working:%v][job:%d]\n", v.id, v.working, len(v.job))
	}
}

//=====================================================
type Worker struct {
	id          int
	job         chan *Job
	working     bool
	recoverFunc func()
}

func (w *Worker) work() {
	defer func() {
		fmt.Printf("%v Finished\n", w)
		w.working = false
	}()

	w.working = true
	for {
		select {
		case j := <-w.job:
			func() {
				if w.recoverFunc != nil {
					defer w.recoverFunc()
				}
				j.f()
			}()
			close(j.end)
		}
	}
}
