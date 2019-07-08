package util

import (
	"cell/common/mustang/log"
	"cell/common/mustang/message_queue"
	"cell/common/mustang/message_queue/message_queue_nats"
	"cell/common/mustang/signals"
	"net"
	"runtime"
	"sync"
	"time"
)

type ServiceConfig struct {
	ID           uint64
	BindIP       string
	BindCellPort string
	BindHttpPort string
	BindTcpPort  string
	BindWsPort   string
	BindWsUrl    string
	Group        *sync.WaitGroup
}

type IMsgQueueProcessor interface {
	SetMsgQueInited()
	ProcessSubcribe() bool
	Subscribe(t string, f interface{}) error
	Publish(topic string, msgData interface{})
}

type IService interface {
	Init(*ServiceConfig) bool
	MainLoop() error
	Final() bool
	Main(*ServiceConfig) bool
	ID() uint64
	Name() string
	Type() uint64

	IMsgQueueProcessor
}

type Service struct {
	terminate       bool
	maxFailureDelay time.Duration
	Derived         IService
	ServiceName     string
	ServiceID       uint64
	ServiceType     uint64
	ServiceAddr     string
	IsValided       bool

	msgQueue message_queue.IMsgQueue
}

func (this *Service) Name() string {
	return this.ServiceName
}

func (this *Service) ID() uint64 {
	return this.ServiceID
}

func (this *Service) Type() uint64 {
	return this.ServiceType
}

func (this *Service) Addr() string {
	return this.ServiceAddr
}

func (this *Service) Terminate() bool {
	this.terminate = true
	this.Derived.Final()
	return true
}

func (self *Service) isTerminate() bool {
	return self.terminate
}

func (self *Service) Main(def *ServiceConfig) bool {
	self.IsValided = true

	runtime.GOMAXPROCS(runtime.NumCPU())

	self.maxFailureDelay = 1 * time.Second
	self.terminate = false
	self.msgQueue = message_queue_nats.NewMsgQueue() // 如果要换消息队列九更改这一个地方

	if !self.Derived.Init(def) {
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

		if self.msgQueue != nil && self.msgQueue.InitMsgQue(self.ServiceName) == true {
			self.Derived.ProcessSubcribe()
		}
	}
	log.Info("[服务%s] 结束退出\n", self.ServiceName)
}

//ServiceDef 服务定义 ====================================================
type ServiceDef struct {
	ServiceId   uint64
	ServiceAddr string
	ServiceType uint64
	TcpAddr     string
}

//MsgQueue  消息队列  ====================================================
func (self *Service) GetIMsgQue() message_queue.IMsgQueue {
	return self.msgQueue
}

func (self *Service) SetMsgQueInited() {
	self.GetIMsgQue().SetMsgQueInited()
}

// 在每个具体服务器实现
func (self *Service) ProcessSubcribe() bool {
	return false
}

func (self *Service) Subscribe(t string, f interface{}) error {
	return self.GetIMsgQue().Subscribe(t, f)
}

func (self *Service) Publish(topic string, msgData interface{}) {
	self.GetIMsgQue().Publish(topic, msgData)
}
