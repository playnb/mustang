package worker

import "sync"

//用于调节同时运行的线程数量，并且等待所有线程结束
type WaitAndLimit struct {
	wait     *sync.WaitGroup
	limit    chan bool
	maxLimit int
}

//在开始go之前调用
func (wl *WaitAndLimit) Add() {
	wl.wait.Add(1)
	wl.limit <- true
}
func (wl *WaitAndLimit) AddNum(count int) {
	for count > 0 {
		count--
		wl.wait.Add(1)
		wl.limit <- true
	}
}

//go的函数中结束时（defer）调用
func (wl *WaitAndLimit) Done() {
	wl.wait.Done()
	<-wl.limit
}

//主线程等待所有Done
func (wl *WaitAndLimit) Wait() {
	wl.wait.Wait()
}

func (wl *WaitAndLimit) MaxLimit() int {
	return wl.maxLimit
}

func NewWaitAndLimit(limit int) *WaitAndLimit {
	w := &WaitAndLimit{
		wait:     &sync.WaitGroup{},
		limit:    make(chan bool, limit),
		maxLimit: limit,
	}
	return w
}
