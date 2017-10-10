package until

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/signals"
	"net"
	"runtime"
	"time"
)

type IService interface {
	Init() bool
	MainLoop() error
	Final() bool
	Main() bool
}

type Service struct {
	terminate       bool
	maxFailureDelay time.Duration
	Derived         IService
	ServiceName     string
	IsValided       bool
}

func (this *Service) Name() string {
	return this.ServiceName
}

func (this *Service) Terminate() bool {
	this.terminate = true
	this.Derived.Final()
	return true
}

func (self *Service) isTerminate() bool {
	return self.terminate
}

func (self *Service) Main() bool {
	self.IsValided = true

	runtime.GOMAXPROCS(runtime.NumCPU())

	self.maxFailureDelay = 1 * time.Second
	self.terminate = false

	if !self.Derived.Init() {
		return false
	}

	//TODO 这么玩确定可以有多个IService,同时运行?
	exitChan := signals.Start()
	signals.AddSubscriber(signals.EXIT, self.Terminate)

	go self.RunLoop()

	//TODO 让出CPU?有这个必要吗?
	runtime.Gosched()
	<-exitChan

	self.IsValided = false
	return true
}

func (self *Service) RunLoop() {
	/*
		defer func() {
			if err := recover(); err != nil {
				log.Error("[异常] ", err, "\n", string(debug.Stack()))
			}
		}()
	*/
	log.Trace("%s 服务启动", self.ServiceName)

	//defer log.PrintPanicStack()
	defer func() {
		if err := recover(); err != nil {
			log.PrintPanicError(err)

			//log.Trace("%s 服务panic 重新调用RunLoop", self.ServiceName)
			//go self.RunLoop()
		}
	}()

	var tempDelay time.Duration // how long to sleep on accept failure

	for !self.isTerminate() {
		//TODO 搞个Tiker控制得了
		time.Sleep(time.Millisecond * 50)

		//log.Trace("%s ==========", self.ServiceName)
		err := self.Derived.MainLoop()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if tempDelay > self.maxFailureDelay {
					tempDelay = self.maxFailureDelay
				}
				log.Error("Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			//return err
		}
	}
	log.Info("[服务%s] 结束退出\n", self.ServiceName)
}

//ServiceDef 服务定义
type ServiceDef struct {
	ServiceId   uint64
	ServiceAddr string
	ServiceType uint64
	TcpAddr     string
}
