package network

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/until"
	"time"
)

type TCPAgent struct {
	name           string
	conn           *TCPConn
	msgProcessor   IMsgProcessor
	handlerMessage HandlerMessageFunc
	userData       interface{}

	agentMsgChan chan *AgentMessage
	tcpMsgChan   chan []byte

	MethodAgent *until.MethodAgent
	DebugMe     bool
	timerAgent  ITimerAgent
}

func (a *TCPAgent) SetDebug(debug bool) {
	a.DebugMe = debug
}
func (a *TCPAgent) GetDebug() bool {
	return a.DebugMe
}
func (a *TCPAgent) SetTimerAgent(timerAgent ITimerAgent) {
	a.timerAgent = timerAgent
}

//TODO 这里可以修改一下,网络部分每个连接至少3个gorountinue
//1.负责发送消息 2.负责接收消息 3.处理消息(这里是处理消息的)
//现在是两个  1个是发送线程(没有逻辑) 1个是接收线程(逻辑都在这)

//Run 阻塞运行,接收消息
func (a *TCPAgent) Run(agent IAgent) {
	a.agentMsgChan = make(chan *AgentMessage, 100)
	a.tcpMsgChan = make(chan []byte, 100)
	go func() {
		defer log.PrintPanicStack()

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
			case <-a.conn.exitChan:
				a.Debug("%s 主动接收TCPAgent.Run.Logic结束", a.conn)
				//TODO: agentMsgChan,tcpMsgChan,没有关闭哦.....
				return
			case <-ts.C:
				if a.timerAgent != nil {
					a.timerAgent.SecondTimerFunc()
				}
			case <-tm.C:
				if a.timerAgent != nil {
					a.timerAgent.MinuteTimerFunc()
				}
			case <-tfm.C:
				if a.timerAgent != nil {
					a.timerAgent.FiveMinuteTimerFunc()
				}
			case <-th.C:
				if a.timerAgent != nil {
					a.timerAgent.HourTimerFunc()

					// if time.Now().Format("15") == "00" {
					// 	a.ZeroTimerFunc()
					// }
				}
			case msg, ok := <-a.agentMsgChan:
				if ok {
					if a.MethodAgent != nil {
						msg.OutChan <- a.MethodAgent.Call(a.userData, msg.MsgName, msg.InData)
					} else {
						msg.OutChan <- nil
					}
				}
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

			a.tcpMsgChan <- data
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
}

func (a *TCPAgent) Call(name string, in interface{}) interface{} {
	agm := &AgentMessage{
		MsgName: name,
		InData:  in,
		OutChan: make(chan interface{}),
	}
	a.agentMsgChan <- agm
	return <-agm.OutChan
}

func (a *TCPAgent) GetCloseFlag() bool {
	return a.conn.closeFlag
}

func (a *TCPAgent) SetCloseFlag() {
	a.conn.closeFlag = true
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
func (a *TCPAgent) WriteMsg(msg interface{}) error {
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
		data := msg.([]byte)
		if data != nil {
			return a.conn.WriteMsg(data)
		} else {
			a.Warning("发送消息没有合适的处理器")
		}
	}
	return nil
}

func (a *TCPAgent) WriteMsgWithoutPack(data []byte) error {
	return a.conn.WriteMsg(data)
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

func (a *TCPAgent) SetConn(conn interface{}) {
	a.conn = conn.(*TCPConn)
}

func (a *TCPAgent) Name() string {
	return a.name
}

func (a *TCPAgent) SetName(name string) {
	a.name = name
}
