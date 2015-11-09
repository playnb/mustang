package network

import (
	"TestGo/mustang/log"
	"github.com/golang/protobuf/proto"
	"reflect"
)

type WSAgent struct {
	name              string
	conn              *WSConn
	protobufProcessor IProtobufProcessor
	userData          interface{}
}

func (a *WSAgent) Run() {
	a.Trace("run...")
	for {
		if a.conn == nil {
			a.Error("WSAgent 连接为nil")
			break
		}
		data, err := a.conn.ReadMsg()
		if err != nil {
			a.Debug("读取消息错误: %v", err)
			break
		}

		if a.protobufProcessor != nil {
			_, err = a.protobufProcessor.Handler(a, data, a.userData)
			if err != nil {
				a.Debug("protobuf handle msg error %v", err)
				//WHY: 解析错误需要跳出循环吗?
				//break
			}
		}
	}
}

func (a *WSAgent) OnClose() {
	log.Trace("WSAgent 退出")
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
		a.Error("发送消息没有合适的处理器")
	}
}

func (a *WSAgent) Close() {
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

func (a *WSAgent) UserData() interface{} {
	return a.userData
}

func (a *WSAgent) SetUserData(data interface{}) {
	a.userData = data
}

func (a *WSAgent) ProtobufProcessor() IProtobufProcessor {
	return a.protobufProcessor
}

func (a *WSAgent) SetProtobufProcessor(processor IProtobufProcessor) {
	a.protobufProcessor = processor
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
