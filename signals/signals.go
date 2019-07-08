package signals

import (
	"github.com/playnb/mustang/log"
	"container/list"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

/*
处理Linux的信号
*/

type Signal int

const (
	_ Signal = iota
	EXIT
	RELOAD
	STATUS
)

type HandlerFunc func() bool

var signalData *signalObj = &signalObj{
	publish: make(map[Signal]*list.List),
}

type signalObj struct {
	publish map[Signal]*list.List
	mtx     sync.Mutex
}

var exit_chan chan bool = make(chan bool, 1)

func ExitChan() chan bool {
	return exit_chan
}

//启动监听
func Start() chan bool {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		defer signal.Stop(ch)
		for sig := range ch {
			log.Info("[服务] 收到信号:", sig)
			notify(sig)
		}
	}()
	return exit_chan
}

func AddSubscriber(sig_type Signal, f HandlerFunc) {
	signalData.mtx.Lock()
	defer signalData.mtx.Unlock()

	funcs, ok := signalData.publish[sig_type]
	if funcs == nil || !ok {
		funcs = list.New()
	}
	funcs.PushBack(f)
	signalData.publish[sig_type] = funcs
}

func SendSignalTimeOut(delay time.Duration, sigs ...syscall.Signal) {
	time.AfterFunc(delay, func() {
		if proc, err := os.FindProcess(os.Getegid()); err == nil {
			for _, sig := range sigs {
				proc.Signal(sig)
			}
		}
	})
}

func SendSignal(sigs ...syscall.Signal) {
	if proc, err := os.FindProcess(os.Getegid()); err == nil {
		for _, sig := range sigs {
			proc.Signal(sig)
		}
	}
}

func notify(sig os.Signal) bool {
	sig_type := EXIT
	switch sig {
	case syscall.SIGHUP: //重新读配置文件
		sig_type = RELOAD
	case syscall.Signal(29): //获取服务状态
		sig_type = STATUS
	}

	runHandler(sig_type)

	if sig_type == EXIT {
		log.Info("[退出] 收到退出指令")
		time.Sleep(time.Second * 1)
		close(exit_chan)
		return false
	}
	return false
}

func runHandler(sig Signal) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("[异常] ", err, "\n", string(debug.Stack()))
			log.Flush()
		}
	}()

	var handler HandlerFunc
	if funcs, ok := signalData.publish[sig]; funcs != nil && ok {
		for f := funcs.Front(); f != nil; f = f.Next() {
			handler = f.Value.(HandlerFunc)
			handler()
		}
	}
}
