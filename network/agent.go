package network

//	"reflect"

// "github.com/golang/protobuf/proto"

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
type MsgHandler func(agent IAgent, message interface{}, msgIndex uint16, data interface{})
type HandlerMessageFunc func(reciver IAgent, msg *AgentMessage) (bool, error)

/*
//protobuf消息处理器
type IProtobufProcessor interface {
	//注册消息及处理函数
	Register(msg proto.Message, msgHandler MsgHandler)
	//调用消息映射的处理函数
	Handler(reciver IAgent, data []byte, clientData interface{}) (bool, error)
	//打包消息
	PackMsg(msg proto.Message) ([]byte, error)
}
*/

type IMsgProcessor interface {
	//调用消息映射的处理函数
	Handler(reciver IAgent, data []byte, clientData interface{}) (bool, error)
	Handler2(reciver IAgent, msgIndex uint16, msgData interface{}) (bool, error)
	//打包消息
	PackMsg(msg interface{}) ([]byte, error)
	//注册处理函数
	Register(sample interface{}, msgHandler MsgHandler)
}

//处理链接的代理对象
type IAgent interface {
	SetConn(conn interface{})
	SetTimerAgent(ITimerAgent)

	//阻塞运行,接收消息
	Run(agent IAgent)
	//发送消息
	WriteMsg(msg interface{}) error
	WriteMsgWithoutPack(data []byte) error

	SetUserData(data interface{})

	CloseFunc()
	ConnectFunc()

	SetDebug(bool)
	GetDebug() bool
	GetCloseFlag() bool
	Name() string

	Call(name string, in interface{}) interface{}
}

type DefaultAgent struct {
}

func (agent *DefaultAgent) GetCloseFlag() bool                           { return false }
func (agent *DefaultAgent) Call(name string, in interface{}) interface{} { return nil }
func (agent *DefaultAgent) WriteMsg(msg interface{}) error               { return nil }
func (agent *DefaultAgent) WriteMsgWithoutPack(data []byte) error        { return nil }
func (agent *DefaultAgent) OnClose()                                     {}
func (agent *DefaultAgent) SetDebug(debug bool)                          {}
func (agent *DefaultAgent) GetDebug() bool                               { return false }
func (agent *DefaultAgent) SetConn(interface{})                          {}
func (agent *DefaultAgent) Name() string                                 { return "DefaultAgent" }
func (agent *DefaultAgent) SetTimerAgent(ITimerAgent)                    {}
func (agent *DefaultAgent) SetUserData(data interface{})                 {}
