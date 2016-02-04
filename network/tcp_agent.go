package network

import (
	"github.com/golang/protobuf/proto"
	"github.com/playnb/mustang/log"
	//	"reflect"
)

type TCPAgent struct {
	name              string
	conn              *TCPConn
	protobufProcessor IProtobufProcessor
	userData          interface{}
}

//阻塞运行,接收消息
func (a *TCPAgent) Run() {
	a.Trace("run...")
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			a.Error("Client readMsg error: %v", err)
			break
		}

		if a.protobufProcessor != nil {
			_, err := a.protobufProcessor.Handler(a, data, nil)
			if err != nil {
				a.Error("Client handler error: %v", err)
			}
		}
	}
}

//发送消息
func (a *TCPAgent) WriteMsg(msg interface{}) {
	if a.protobufProcessor != nil {
		data, err := a.protobufProcessor.PackMsg(msg.(proto.Message))
		if err != nil {
			log.Debug("protobuf pack msg error %v", err)
			return
		}
		a.conn.WriteMsg(data)
	}
}

//关闭时调用
func (a *TCPAgent) OnClose() {
	log.Trace("TCPClientAgent exit")
}

func (a *TCPAgent) description() string {
	return "[" + a.name + "] "
}

func (a *TCPAgent) Debug(format string, d ...interface{}) {
	log.Debug(a.description()+format, d...)
}

func (a *TCPAgent) Trace(format string, d ...interface{}) {
	log.Trace(a.description()+format, d...)
}

func (a *TCPAgent) Error(format string, d ...interface{}) {
	log.Error(a.description()+format, d...)
}

func (a *TCPAgent) Fatal(format string, d ...interface{}) {
	log.Fatal(a.description()+format, d...)
}

func (a *TCPAgent) UserData() interface{} {
	return a.userData
}

func (a *TCPAgent) SetUserData(data interface{}) {
	a.userData = data
}

func (a *TCPAgent) ProtobufProcessor() IProtobufProcessor {
	return a.protobufProcessor
}

func (a *TCPAgent) SetProtobufProcessor(processor IProtobufProcessor) {
	a.protobufProcessor = processor
}

func (a *TCPAgent) SetConn(conn *TCPConn) {
	a.conn = conn
}

func (a *TCPAgent) Name() interface{} {
	return a.name
}

func (a *TCPAgent) SetName(name string) {
	a.name = name
}
