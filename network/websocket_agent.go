package network

import (
	"github.com/playnb/mustang/log"
	"errors"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

//WSAgent websocket的agent
type WSAgent struct {
	name           string
	conn           *WSConn
	msgProcessor   IMsgProcessor
	handlerMessage HandlerMessageFunc
	innerMsgChan   chan []byte
	agentMsgChan   chan *AgentMessage
	msgChan        chan []byte
	UserData       interface{}
	needClose      chan bool
}

func (a *WSAgent) String() string {
	return "[" + a.name + "] "
}

//Close 主动关闭 WSAgent
func (a *WSAgent) Close() {
	log.Debug("%s 主动关闭 WSAgent", a)
	a.conn.Close()
}

//WriteMsg 发送数据
func (a *WSAgent) WriteMsg(msg interface{}) error {
	if a.msgProcessor != nil {
		// protobuf
		data, err := a.msgProcessor.PackMsg(msg.(proto.Message))
		if err != nil {
			log.Error("%s marshal protobuf %v error: %v", a, reflect.TypeOf(msg), err)
			return errors.New("marshal protobuf error")
		}
		return a.conn.WriteMsg(data)
	}
	data := msg.([]byte)
	if data != nil {
		return a.conn.WriteMsg(data)
	}
	log.Error("%s 发送消息没有合适的处理器", a)
	return errors.New("发送消息没有合适的处理器")
}

func (a *WSAgent) WriteMsgWithoutPack(data []byte) error {
	return a.conn.WriteMsg(data)
}

//Run 运行Agent
func (a *WSAgent) Run(agent IAgent) {
	a.needClose = make(chan bool)
	a.msgChan = make(chan []byte, 1024)
	a.agentMsgChan = make(chan *AgentMessage, 100)

	go func() {
		defer log.PrintPanicStack()
		for {
			if a.conn == nil {
				log.Error("%s WSAgent 连接为nil", a)
				close(a.needClose)
				return
			}
			data, err := a.conn.ReadMsg()
			if err != nil {
				log.Debug("%s  读取消息错误: %v", a, err)
				close(a.needClose)
				return
			}

			a.msgChan <- data
		}
	}()

	for {
		select {
		case data := <-a.msgChan:
			if a.msgProcessor != nil {
				_, err := a.msgProcessor.Handler(agent, data, a.UserData)
				if err != nil {
					log.Error("[protobufProcessor] %s protobuf handle msg error %v", a, err)
				}
			} else {

			}
		case ok, _ := <-a.needClose:
			if ok == false {
				///要退出了
			}
		}
	}
}

//Call 异步调用方法
func (a *WSAgent) Call(name string, in interface{}) interface{} {
	agm := &AgentMessage{
		MsgName: name,
		InData:  in,
		OutChan: make(chan interface{}),
	}
	a.agentMsgChan <- agm
	return <-agm.OutChan
}

func (a *WSAgent) GetDebug() bool {
	return false
}
func (a *WSAgent) SetDebug(bool) {
}

func (a *WSAgent) Name() string {
	return a.String()
}

func (a *WSAgent) SetConn(conn interface{}) {

}
