package gate

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/network"
	//"github.com/playnb/mustang/network/protobuf"
	//"github.com/golang/protobuf/proto"
	//"reflect"
	"time"
)

type NewProcessorFunc func(network.IAgent) network.IProtobufProcessor

type WSGate struct {
	Name              string
	Addr              string
	MaxConnNum        int           //最大连接数量
	PendingWriteNum   int           //最大消息队列数量
	HTTPTimeout       time.Duration //http超时时间(ms)
	ProtobufProcessor network.IProtobufProcessor
	HandlerMessage    network.HandlerMessageFunc
	//agent             *WSGateAgent
	agents map[*WSGateAgent]bool
}

func (gate *WSGate) Run(closeSig chan bool) {
	server := &network.WSServer{}
	gate.agents = make(map[*WSGateAgent]bool)
	server.Addr = gate.Addr
	server.MaxConnNum = gate.MaxConnNum
	server.HTTPTimeout = gate.HTTPTimeout
	server.PendingWriteNum = gate.PendingWriteNum
	server.NewAgent = func(conn *network.WSConn) network.IAgent {
		a := new(WSGateAgent)
		a.SetConn(conn)
		a.SetName("WSGateAgent")
		a.SetProtobufProcessor(gate.ProtobufProcessor)
		a.SetHandlerMessageFunc(gate.HandlerMessage)
		a.gate = gate
		gate.agents[a] = true
		//gate.agent = a
		return a
	}
	server.Start()
	//等待close信号
	<-closeSig
	server.Close()
}

func (gate *WSGate) Dump() {
	log.Trace("%s 当前连接数 %u", gate.Name, len(gate.agents))
}

/*
func (gate *WSGate) WriteMsg(msg interface{}) {
	if gate.agent != nil {
		gate.agent.WriteMsg(msg)
	}
}
*/
type WSGateAgent struct {
	network.WSAgent
	gate *WSGate
}

func (a *WSGateAgent) OnClose() {
	a.gate.agents[a] = false
	delete(a.gate.agents, a)
	log.Trace("WSGateAgent 退出")
}
