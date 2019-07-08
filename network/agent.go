package network

import (
	"github.com/playnb/mustang/util"
	"net"
	"sync/atomic"
)

//	"reflect"

// "github.com/golang/protobuf/proto"

//为TCP打包预留的消息头大小
const BufHeadOffset = 4

//ITimerAgent 定时器Agent
type ITimerAgent interface {
	SecondTimerFunc()
	MinuteTimerFunc()
	FiveMinuteTimerFunc()
	HourTimerFunc()
	ZeroTimerFunc()
}

//AgentMessage agent接受本进程的消息
type AgentMessage struct {
	MsgName string
	InData  interface{}
	OutChan chan interface{}
}

//消息类型与处理函数的映射关系
//agent 接受的消息的对象
//msgType 消息类型
//message 消息数据
//data 附加数据(data[0]原始消息数据 data[1]userData)
type MsgHandler func(agent SampleAgent, message interface{}, msgIndex uint16, data interface{})
type HandlerMessageFunc func(receiver IAgent, msg *AgentMessage) (bool, error)

/*
//protobuf消息处理器
type IProtobufProcessor interface {
	//注册消息及处理函数
	Register(msg proto.Message, msgHandler MsgHandler)
	//调用消息映射的处理函数
	Handler(receiver IAgent, data []byte, clientData interface{}) (bool, error)
	//打包消息
	PackMsg(msg proto.Message) ([]byte, error)
}
*/

type ICanRegisterHandler interface {
	Register(sample interface{}, msgHandler MsgHandler)
}

type IMsgProcessor interface {
	//调用消息映射的处理函数
	Handler(receiver SampleAgent, data *util.BuffData, clientData interface{}) (bool, error)
	//打包消息
	PackMsg(msg MarshalToProto) (*util.BuffData, error)
	//注册处理函数
	Register(sample interface{}, msgHandler MsgHandler)
}

type SampleAgent interface {
	WriteMsg(msg MarshalToProto) error
	WriteMsgWithoutPack(data *util.BuffData) error

	//SetOnDataHandle(func(*util.BuffData))

	GetConnExitChan() chan bool //退出的标识Chan

	UserData() interface{}
	SetUserData(data interface{})

	GetUniqueID() uint64
}

//处理链接的代理对象
type IAgent interface {
	SetConn(conn IConn)
	SetTimerAgent(ITimerAgent)

	//接受的消息的Chan
	GetMsgChan() chan *util.BuffData

	//阻塞运行,接收消息
	Run(agent IAgent)
	//发送消息
	SampleAgent

	CloseFunc()
	ConnectFunc()

	SetDebug(bool)
	GetDebug() bool
	GetCloseFlag() bool
	Name() string
	SetName(name string)

	RemoteAddr() net.Addr
}

type DefaultAgent struct {
}

func (agent *DefaultAgent) GetCloseFlag() bool                            { return false }
func (agent *DefaultAgent) Call(name string, in interface{}) interface{}  { return nil }
func (agent *DefaultAgent) WriteMsg(msg MarshalToProto) error             { return nil }
func (agent *DefaultAgent) WriteMsgWithoutPack(data *util.BuffData) error { return nil }
func (agent *DefaultAgent) OnClose()                                      {}
func (agent *DefaultAgent) SetDebug(debug bool)                           {}
func (agent *DefaultAgent) GetDebug() bool                                { return false }
func (agent *DefaultAgent) SetConn(interface{})                           {}
func (agent *DefaultAgent) Name() string                                  { return "DefaultAgent" }
func (agent *DefaultAgent) SetTimerAgent(ITimerAgent)                     {}
func (agent *DefaultAgent) SetUserData(data interface{})                  {}
func (agent *DefaultAgent) RemoteAddr() net.Addr                          { return nil }
func (agent *DefaultAgent) Do(name string, in interface{})                {}
func (agent *DefaultAgent) SetName(name string)                           {}
func (agent *DefaultAgent) ConnectTempID() uint64                         { return 0 }

var connect_sequence_id uint64
//生成唯一的连接ID
func get_connect_sequence_id() uint64 {
	return atomic.AddUint64(&connect_sequence_id, 1)
}
