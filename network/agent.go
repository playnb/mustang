package network

import (
	"github.com/golang/protobuf/proto"
	"reflect"
)

//消息类型与处理函数的映射关系
//agent 接受的消息的对象
//msgType 消息类型
//message 消息数据
//data 附加数据(data[0]原始消息数据 data[1]userData)
type MsgHandler func(agent IAgent, msgType reflect.Type, message interface{}, data []interface{})

//protobuf消息处理器
type IProtobufProcessor interface {
	//注册消息及处理函数
	Register(msg proto.Message, msgHandler MsgHandler)
	//调用消息映射的处理函数
	Handler(reciver IAgent, data []byte, clientData interface{}) (bool, error)
	//打包消息
	PackMsg(msg proto.Message) ([]byte, error)
}

//处理链接的代理对象
type IAgent interface {
	//阻塞运行,接收消息
	Run()
	//发送消息
	WriteMsg(msg interface{})
	//关闭时调用
	OnClose()
}
