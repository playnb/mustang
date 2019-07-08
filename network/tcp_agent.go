package network

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/util"
	"fmt"
	"net"
	"sync"
	"time"
)

type TCPAgent struct {
	connectTempID  uint64
	name           string
	conn           *TCPConn
	msgProcessor   IMsgProcessor
	handlerMessage HandlerMessageFunc
	userData       interface{}

	agentMsgChan    chan *AgentMessage
	tcpMsgChan      chan *util.BuffData
	TcpMsgChanCount uint64

	MethodAgent *util.MethodAgent
	DebugMe     bool
	timerAgent  ITimerAgent

	UserThread      bool
	CanDropMessage  bool
	ProcessDirectly bool

	allCallsRetChan    map[string]chan interface{}
	workingGorountinue bool
	workingMutex       sync.Mutex
}

func (a *TCPAgent) SetDebug(debug bool) {
	a.DebugMe = debug
	if debug {
		log.Trace("IAgent调试信息 [%s] msgParser:[%s] tcpMsgChan:[%v]", a, a.conn.msgParser, len(a.tcpMsgChan))
	}
}

func (ia *TCPAgent) GetUserID() uint64 {
	return 0
}
func (ia *TCPAgent) SetUserID(uid uint64) {
}
func (ia *TCPAgent) GetUniqueID() uint64 {
	return ia.connectTempID
}

func (a *TCPAgent) SetOnDataHandle(func(*util.BuffData)) {}

func (a *TCPAgent) GetDebug() bool {
	return a.DebugMe
}
func (a *TCPAgent) SetTimerAgent(timerAgent ITimerAgent) {
	a.timerAgent = timerAgent
}

func (a *TCPAgent) GetAgentMsgChan() chan *AgentMessage {
	if a.UserThread == false {
		return nil
	}
	return a.agentMsgChan
}
func (a *TCPAgent) GetMsgChan() chan *util.BuffData {
	if a.UserThread == false {
		return nil
	}
	return a.tcpMsgChan
}

//Conn推出的Chan
func (a *TCPAgent) GetConnExitChan() chan bool {
	return a.conn.GetExitChan()
}
func (a *TCPAgent) Init() {
	if a.connectTempID == 0 {
		a.connectTempID = get_connect_sequence_id()
	}
	if a.agentMsgChan == nil {
		a.agentMsgChan = make(chan *AgentMessage, 1024)
	}
	if a.tcpMsgChan == nil {
		a.tcpMsgChan = make(chan *util.BuffData, 1024)
	}
	if a.allCallsRetChan == nil {
		a.allCallsRetChan = make(map[string]chan interface{})
	}
}
func (a *TCPAgent) checkAgentMsgChanCap() {
	l := len(a.agentMsgChan)
	c := cap(a.agentMsgChan)
	if l > c*2/3 {
		log.Trace("消息累积 agentMsgChan %d", c)
	}
}

//TODO 这里可以修改一下,网络部分每个连接至少3个gorountinue
//1.负责发送消息 2.负责接收消息 3.处理消息(这里是处理消息的)
//现在是两个  1个是发送线程(没有逻辑) 1个是接收线程(逻辑都在这)

//Run 阻塞运行,接收消息
func (a *TCPAgent) Run(agent IAgent) {
	a.Init()
	var waitGroup sync.WaitGroup
	if a.UserThread == false {
		waitGroup.Add(1)
		go func() {
			defer log.PrintPanicStack()
			defer func() {
				a.workingMutex.Lock()
				a.Debug("%s 主动接收TCPAgent.Run.Logic结束 Call:%d", a.conn, len(a.allCallsRetChan))
				a.workingGorountinue = false
				//a.allCallsRetChan.Range(func(key string, c chan interface{}) bool {
				for _, c := range a.allCallsRetChan {
					close(c)
					//return true
				}
				a.workingMutex.Unlock()
				waitGroup.Done()
			}()

			a.workingGorountinue = true
			ts := time.NewTicker(time.Second)
			tm := time.NewTicker(time.Minute)
			tfm := time.NewTicker(time.Minute * 5)
			th := time.NewTicker(time.Hour)
			defer ts.Stop()
			defer tm.Stop()
			defer tfm.Stop()
			defer th.Stop()

			//FIX:: 这么分流以后就消息没有顺序性了
			if a.TcpMsgChanCount > 0 && false {
				for i := uint64(0); i < a.TcpMsgChanCount; i++ {
					waitGroup.Add(1)
					go func(index uint64) {
						workerName := fmt.Sprintf("tcpMsgChan:cell%d", index)
						defer waitGroup.Done()
						defer log.PrintPanicStack()
						for {
							select {
							case <-a.conn.exitChan:
								//TODO: agentMsgChan,tcpMsgChan,没有关闭哦.....
								return
							case data, ok := <-a.tcpMsgChan:
								__ft := util.NewFunctionTime(workerName, 5)
								if ok {
									if a.msgProcessor != nil {
										_, err := a.msgProcessor.Handler(agent, data, a.UserData)
										if err != nil {
											a.Debug("%s msgProcessor error: %v", a.conn, err)
										}
									}
								}
								__ft.End()
							}
						}
					}(i)
				}
			}

			for {
				select {
				case <-a.conn.exitChan:
					//TODO: agentMsgChan,tcpMsgChan,没有关闭哦.....
					return
				case <-ts.C:
					if a.timerAgent != nil {
						__ft := util.NewFunctionTime("SecondTimerFunc:"+a.Name(), 5)
						a.timerAgent.SecondTimerFunc()
						__ft.End()
					}
				case <-tm.C:
					if a.timerAgent != nil {
						__ft := util.NewFunctionTime("MinuteTimerFunc:"+a.Name(), 5)
						a.timerAgent.MinuteTimerFunc()
						__ft.End()
					}
				case <-tfm.C:
					if a.timerAgent != nil {
						__ft := util.NewFunctionTime("FiveMinuteTimerFunc:"+a.Name(), 5)
						a.timerAgent.FiveMinuteTimerFunc()
						__ft.End()
					}
				case <-th.C:
					if a.timerAgent != nil {
						__ft := util.NewFunctionTime("HourTimerFunc:"+a.Name(), 5)
						a.timerAgent.HourTimerFunc()
						__ft.End()
						// if time.Now().Format("15") == "00" {
						// 	a.ZeroTimerFunc()
						// }
					}
				case msg, ok := <-a.agentMsgChan:
					__ft := util.NewFunctionTime("agentMsgChan:"+a.Name()+":"+msg.MsgName, 5)
					if ok {
						if a.MethodAgent != nil {
							if msg.OutChan != nil {
								msg.OutChan <- a.MethodAgent.Call(a.userData, msg.MsgName, msg.InData)
							} else {
								a.MethodAgent.Call(a.userData, msg.MsgName, msg.InData)
							}
						} else {
							if msg.OutChan != nil {
								msg.OutChan <- nil
							}
						}
					}
					__ft.End()
				case data, ok := <-a.tcpMsgChan:
					if ok {
						if a.msgProcessor != nil {
							_, err := a.msgProcessor.Handler(agent, data, a.UserData)
							if err != nil {
								a.Debug("%s msgProcessor error: %v", a.conn, err)
							}
						}
					}
				}
			}
		}()
	} else {
		a.workingGorountinue = true
	}

	for {
		select {
		case <-a.conn.exitChan:
			a.Debug("%s 主动接收TCPAgent.Run结束", a.conn)
			return
		default:
			data, err := a.conn.ReadMsg()
			if err != nil {
				a.Error("%s 被动接收 Client readMsg error: %v 主动:%v", a.conn, err, a.conn.closeFlag)
				return
			}
			if a.GetDebug() {
				log.Trace("IAgent调试信息 原始数据 %v %v", a, data)
			}

			if a.ProcessDirectly {
				if a.msgProcessor != nil {
					_, err := a.msgProcessor.Handler(agent, data, a.UserData)
					if err != nil {
						a.Debug("%s msgProcessor error: %v", a.conn, err)
					}
				}
			} else {
				a.tcpMsgChan <- data
			}
			/*
				if a.msgProcessor != nil {
					_, err := a.msgProcessor.Handler(agent, data, a.UserData)
					if err != nil {
						a.Debug("%s msgProcessor error: %v", a.conn, err)
					}
				}
			*/
		}
	}
	waitGroup.Wait()
}

func (a *TCPAgent) Call(name string, in interface{}) interface{} {
	agm := &AgentMessage{
		MsgName: name,
		InData:  in,
		OutChan: make(chan interface{}),
	}
	key := fmt.Sprintf("%p", agm.OutChan)

	a.workingMutex.Lock()
	//a.allCallsRetChan.Store(key, agm.OutChan)
	a.allCallsRetChan[key] = agm.OutChan
	a.workingMutex.Unlock()

	if a.UserThread == false {
		if a.workingGorountinue == false {
			log.Debug("==Call %s 的时候 对象%s已经关闭(前置)", name, a)
			return a.MethodAgent.Call(a.userData, agm.MsgName, agm.InData)
			//return nil //Call的对象已经关闭了
		}
	}

	a.checkAgentMsgChanCap()
	a.agentMsgChan <- agm
	ret, ok := <-agm.OutChan

	a.workingMutex.Lock()
	//a.allCallsRetChan.Delete(key)
	delete(a.allCallsRetChan, key)
	a.workingMutex.Unlock()

	if ok == false {
		log.Debug("==Call %s 的时候 对象%s已经关闭(后置)", name, a)
		return nil //Call的对象已经关闭了
	}
	return ret
}

func (a *TCPAgent) Do(name string, in interface{}) {
	agm := &AgentMessage{
		MsgName: name,
		InData:  in,
		OutChan: nil,
	}
	a.checkAgentMsgChanCap()
	a.agentMsgChan <- agm
}

func (a *TCPAgent) GetCloseFlag() bool {
	return a.conn.closeFlag
}

func (a *TCPAgent) SetCloseFlag() {
	a.conn.closeFlag = true
}

func (agent *TCPAgent) RemoteAddr() net.Addr {
	return agent.conn.RemoteAddr()
}

//Terminate 关闭连接("请求")
func (a *TCPAgent) Terminate() {
	if a == nil {
		return
	}
	log.Debug("%s TCPAgent Terminate, closeFlag:%v", a.conn, a.conn.closeFlag)
	a.conn.Terminate()
}

//WriteMsg 发送消息
func (a *TCPAgent) WriteMsg(msg MarshalToProto) error {
	if a.msgProcessor != nil {
		data, err := a.msgProcessor.PackMsg(msg)
		if err != nil {
			log.Warning("protobuf pack msg error %v", err)
			return err
		}
		//if a.DebugMe {
		//	a.Warning("===发送数据 %v", data)
		//}
		return a.conn.WriteMsg(data)
	} else {
		//data := msg.([]byte)
		//if data != nil {
		//	return a.conn.WriteMsg(data)
		//} else {
		a.Warning("发送消息没有合适的处理器")
		//}
	}
	return nil
}

func (a *TCPAgent) WriteMsgWithoutPack(data *util.BuffData) error {
	//buf := TempBuffData(data,4)
	if a.ProcessDirectly {
		return a.conn.WriteMsgDirectly(data)
	} else {
		return a.conn.WriteMsg(data)
	}
}

func (a *TCPAgent) String() string {
	return "[" + a.name + ":" + a.conn.String() + "] "
}

func (a *TCPAgent) Debug(format string, d ...interface{}) {
	log.Debug(a.String()+format, d...)
}

func (a *TCPAgent) Trace(format string, d ...interface{}) {
	log.Trace(a.String()+format, d...)
}

func (a *TCPAgent) Warning(format string, d ...interface{}) {
	log.Warning(a.String()+format, d...)
}

func (a *TCPAgent) Error(format string, d ...interface{}) {
	log.Error(a.String()+format, d...)
}

func (a *TCPAgent) Fatal(format string, d ...interface{}) {
	log.Fatal(a.String()+format, d...)
}

func (a *TCPAgent) UserData() interface{} {
	return a.userData
}

func (a *TCPAgent) SetUserData(data interface{}) {
	a.userData = data
}

func (a *TCPAgent) MsgProcessor() IMsgProcessor {
	return a.msgProcessor
}

func (a *TCPAgent) SetMsgProcessor(processor IMsgProcessor) {
	a.msgProcessor = processor
}

func (a *TCPAgent) SetHandlerMessageFunc(handlerMessage HandlerMessageFunc) {
	a.handlerMessage = handlerMessage
}

func (a *TCPAgent) SetConn(conn IConn) {
	a.conn = conn.(*TCPConn)
	a.conn.canDropMessage = a.CanDropMessage
}

func (a *TCPAgent) Name() string {
	return a.name
}

func (a *TCPAgent) SetName(name string) {
	a.name = name
}
