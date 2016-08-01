package network

import (
	"github.com/golang/protobuf/proto"
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/utils"
	"reflect"
)

type WSAgent struct {
	name              string
	conn              *WSConn
	protobufProcessor IProtobufProcessor
	handlerMessage    HandlerMessageFunc
	innerMsgChan      chan []byte
	UserData          interface{}
	CloseFunc         func()
	needClose         chan bool
}

func (a *WSAgent) DoInnerMsg(data []byte) {
	a.innerMsgChan <- data
}

func (a *WSAgent) Run() {

	a.Trace("run...")
	msgChan := make(chan []byte, 1024)
	a.innerMsgChan = make(chan []byte, 1024)
	a.needClose = make(chan bool)

	go func() {
		defer utils.PrintPanicStack()
		for {
			if a.conn == nil {
				a.Error("WSAgent 连接为nil")
				a.needClose <- true
				return
			}
			data, err := a.conn.ReadMsg()
			if err != nil {
				a.Debug("读取消息错误: %v", err)
				a.needClose <- true
				return
			}

			msgChan <- data
		}
	}()

	for {
		var data []byte
		var err error
		select {
		case data = <-msgChan:
		case data = <-a.innerMsgChan:
		case <-a.needClose:
			{
				log.Error("WSAgent 发生错误")
				a.Close()
				return
			}
		}
		if a.handlerMessage != nil {
			_, err = a.handlerMessage(a, data, a.UserData)
			if err != nil {
				a.Debug("[handlerMessage]protobuf handle msg error %v", err)
				//WHY: 解析错误需要跳出循环吗?
				//break
			}
		} else if a.protobufProcessor != nil {
			_, err = a.protobufProcessor.Handler(a, data, a.UserData)
			if err != nil {
				a.Debug("[protobufProcessor]protobuf handle msg error %v", err)
				//WHY: 解析错误需要跳出循环吗?
				//break
			}
		}
	}
}

func (a *WSAgent) OnClose() {
	log.Trace("WSAgent 退出")
	if a.CloseFunc != nil {
		a.CloseFunc()
	}
}

func (a *WSAgent) WriteMsg(msg interface{}) {
	if a.protobufProcessor != nil {
		// protobuf
		data, err := a.protobufProcessor.PackMsg(msg.(proto.Message))
		if err != nil {
			a.Error("marshal protobuf %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		a.conn.WriteMsg(data)
	} else {
		data := msg.([]byte)
		if data != nil {
			a.conn.WriteMsg(data)
		} else {
			a.Error("发送消息没有合适的处理器")
		}
	}
}

func (a *WSAgent) Close() {
	log.Debug("主动关闭 WSAgent")
	a.conn.Close()
}

func (a *WSAgent) description() string {
	return "[" + a.name + "] "
}

func (a *WSAgent) Debug(format string, d ...interface{}) {
	log.Debug(a.description()+format, d...)
}

func (a *WSAgent) Trace(format string, d ...interface{}) {
	log.Trace(a.description()+format, d...)
}

func (a *WSAgent) Error(format string, d ...interface{}) {
	log.Error(a.description()+format, d...)
}

func (a *WSAgent) Fatal(format string, d ...interface{}) {
	log.Fatal(a.description()+format, d...)
}

func (a *WSAgent) ProtobufProcessor() IProtobufProcessor {
	return a.protobufProcessor
}

func (a *WSAgent) SetProtobufProcessor(processor IProtobufProcessor) {
	a.protobufProcessor = processor
}

func (a *WSAgent) SetHandlerMessageFunc(handlerMessage HandlerMessageFunc) {
	a.handlerMessage = handlerMessage
}

func (a *WSAgent) SetConn(conn *WSConn) {
	a.conn = conn
}

func (a *WSAgent) Name() interface{} {
	return a.name
}

func (a *WSAgent) SetName(name string) {
	a.name = name
}
