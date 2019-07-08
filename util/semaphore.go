package util

//Semaphore 信号量结构
type Semaphore struct {
	sema chan bool
}

//MakeSemaphore 创建一个新的信号量
func MakeSemaphore(max int) *Semaphore {
	sema := &Semaphore{
		sema: make(chan bool, max),
	}
	return sema
}

func (sema *Semaphore) P() {
	sema.sema <- true
}

func (sema *Semaphore) V() {
	<-sema.sema
}
