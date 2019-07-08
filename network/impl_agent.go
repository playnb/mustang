package network

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/util"
	"context"
	"net"
	"sync"
)

type ImplAgent struct {
	name           string
	conn           IConn
	msgProcessor   IMsgProcessor
	handlerMessage HandlerMessageFunc
	userData       interface{}

	tcpMsgChan chan *util.BuffData

	DebugMe    bool
	timerAgent ITimerAgent

	worker WorkerThread

	connectTempID uint64
	userID        uint64

	onData func(*util.BuffData)
}

func (ia *ImplAgent) GetUserID() uint64 {
	return ia.userID
}
func (ia *ImplAgent) SetUserID(uid uint64) {
	ia.userID = uid
}
func (ia *ImplAgent) GetUniqueID() uint64 {
	return ia.connectTempID
}

func (ia *ImplAgent) SetOnDataHandle(f func(*util.BuffData)) {
	ia.onData = f
}

func (ia *ImplAgent) SetDebug(debug bool) {
	ia.DebugMe = debug
}
func (ia *ImplAgent) GetDebug() bool {
	return ia.DebugMe
}

func (ia *ImplAgent) SetWorker(w WorkerThread) {
	ia.worker = w
}

func (ia *ImplAgent) SetTimerAgent(timerAgent ITimerAgent) {
	ia.timerAgent = timerAgent
}
func (ia *ImplAgent) DoJob(f JobFunc, wait bool, ctx context.Context) error {
	if ia.worker != nil {
		return ia.worker.DoJob(f, wait, ctx)
	}
	return nil
}

func (ia *ImplAgent) GetMsgChan() chan *util.BuffData {
	return ia.tcpMsgChan
}

//Conn推出的Chan
func (ia *ImplAgent) GetConnExitChan() chan bool {
	return ia.conn.GetExitChan()
}

func (ia *ImplAgent) Init() {
	if ia.tcpMsgChan == nil {
		ia.tcpMsgChan = make(chan *util.BuffData, 1024)
	}
	ia.connectTempID = get_connect_sequence_id()
}

//TODO 这里可以修改一下,网络部分每个连接至少3个gorountine
//1.负责发送消息 2.负责接收消息 3.处理消息(这里是处理消息的)
//现在是两个  1个是发送线程(没有逻辑) 1个是接收线程(逻辑都在这)

//Run 阻塞运行,接收消息
func (ia *ImplAgent) Run(agent IAgent) {
	ia.Init()

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if ia.worker != nil {
		ia.worker.Go(waitGroup)
	}

	for {
		select {
		case <-ia.conn.GetExitChan():
			ia.Debug("%s 主动接收ImplAgent.Run结束", ia.conn)
			return
		default:
			data, err := ia.conn.ReadMsg()
			if err != nil {
				ia.Error("%s 被动接收 Client readMsg error: %v 主动:%v", ia.conn, err, ia.GetCloseFlag())
				return
			}
			if ia.onData != nil {
				ia.onData(data)
			} else {
				ia.tcpMsgChan <- data
			}
			/*
				if ia.msgProcessor != nil {
					_, err := ia.msgProcessor.Handler(agent, data, ia.UserData)
					if err != nil {
						ia.Debug("%s msgProcessor error: %v", ia.conn, err)
					}
				}
			*/
		}
	}
	waitGroup.Wait()
}

/*
func (ia *ImplAgent) Call(name string, in interface{}) interface{} {
	agm := &AgentMessage{
		MsgName: name,
		InData:  in,
		OutChan: make(chan interface{}),
	}
	key := fmt.Sprintf("%p", agm.OutChan)

	ia.workingMutex.Lock()
	//ia.allCallsRetChan.Store(key, agm.OutChan)
	ia.allCallsRetChan[key] = agm.OutChan
	ia.workingMutex.Unlock()

	if ia.workingGorountinue == false {
		log.Debug("==Call %s 的时候 对象%s已经关闭(前置)", name, ia)
		return ia.MethodAgent.Call(ia.userData, agm.MsgName, agm.InData)
		//return nil //Call的对象已经关闭了
	}

	ia.agentMsgChan <- agm
	ret, ok := <-agm.OutChan

	ia.workingMutex.Lock()
	//ia.allCallsRetChan.Delete(key)
	delete(ia.allCallsRetChan, key)
	ia.workingMutex.Unlock()

	if ok == false {
		log.Debug("==Call %s 的时候 对象%s已经关闭(后置)", name, ia)
		return nil //Call的对象已经关闭了
	}
	return ret
}

func (ia *ImplAgent) Do(name string, in interface{}) {
	agm := &AgentMessage{
		MsgName: name,
		InData:  in,
		OutChan: nil,
	}
	ia.agentMsgChan <- agm
}
*/

func (ia *ImplAgent) GetCloseFlag() bool {
	return ia.conn.GetCloseFlag()
}

func (ia *ImplAgent) SetCloseFlag() {
	ia.conn.SetCloseFlag()
}

func (ia *ImplAgent) RemoteAddr() net.Addr {
	return ia.conn.RemoteAddr()
}

//Terminate 关闭连接("请求")
func (ia *ImplAgent) Terminate() {
	if ia == nil {
		return
	}
	log.Debug("%s ImplAgent Terminate, closeFlag:%v", ia.conn, ia.GetCloseFlag())
	ia.conn.Terminate()
}

//WriteMsg 发送消息
func (ia *ImplAgent) WriteMsg(msg MarshalToProto) error {
	if ia.msgProcessor != nil {
		data, err := ia.msgProcessor.PackMsg(msg)
		if err != nil {
			log.Warning("protobuf pack msg error %v", err)
			return err
		}
		//if ia.DebugMe {
		//	ia.Warning("===发送数据 %v", data)
		//}
		return ia.conn.WriteMsg(data)
	} else {
		//data := msg.([]byte)
		//if data != nil {
		//	return ia.conn.WriteMsg(data)
		//} else {
		ia.Warning("发送消息没有合适的处理器")
		//}
	}
	return nil
}

func (ia *ImplAgent) WriteMsgWithoutPack(data *util.BuffData) error {
	return ia.conn.WriteMsg(data)
}

func (ia *ImplAgent) String() string {
	return "[" + ia.name + ":" + ia.conn.String() + "] "
}

func (ia *ImplAgent) Debug(format string, d ...interface{}) {
	log.Debug(ia.String()+format, d...)
}

func (ia *ImplAgent) Trace(format string, d ...interface{}) {
	log.Trace(ia.String()+format, d...)
}

func (ia *ImplAgent) Warning(format string, d ...interface{}) {
	log.Warning(ia.String()+format, d...)
}

func (ia *ImplAgent) Error(format string, d ...interface{}) {
	log.Error(ia.String()+format, d...)
}

func (ia *ImplAgent) Fatal(format string, d ...interface{}) {
	log.Fatal(ia.String()+format, d...)
}

func (ia *ImplAgent) UserData() interface{} {
	return ia.userData
}

func (ia *ImplAgent) SetUserData(data interface{}) {
	ia.userData = data
}

func (ia *ImplAgent) MsgProcessor() IMsgProcessor {
	return ia.msgProcessor
}

func (ia *ImplAgent) SetMsgProcessor(processor IMsgProcessor) {
	ia.msgProcessor = processor
}

func (ia *ImplAgent) SetHandlerMessageFunc(handlerMessage HandlerMessageFunc) {
	ia.handlerMessage = handlerMessage
}

func (ia *ImplAgent) Name() string {
	return ia.name
}

func (ia *ImplAgent) SetName(name string) {
	ia.name = name
}
