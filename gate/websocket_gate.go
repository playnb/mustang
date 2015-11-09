package gate

import (
	//"TestGo/mustang/log"
	"TestGo/mustang/network"
	"TestGo/mustang/network/protobuf"
	//"github.com/golang/protobuf/proto"
	//"reflect"
	"time"
)

type WSGate struct {
	Addr              string
	MaxConnNum        int           //最大连接数量
	PendingWriteNum   int           //最大消息队列数量
	HTTPTimeout       time.Duration //http超时时间(ms)
	ProtobufProcessor *protobuf.Processor
	agent             *WSGateAgent
}

func (gate *WSGate) Run(closeSig chan bool) {
	server := &network.WSServer{}
	server.Addr = gate.Addr
	server.MaxConnNum = gate.MaxConnNum
	server.HTTPTimeout = gate.HTTPTimeout
	server.PendingWriteNum = gate.PendingWriteNum
	server.NewAgent = func(conn *network.WSConn) network.IAgent {
		a := new(WSGateAgent)
		a.SetConn(conn)
		a.SetName("WSGateAgent")
		a.SetProtobufProcessor(gate.ProtobufProcessor)
		a.gate = gate
		gate.agent = a
		return a
	}
	server.Start()
	//等待close信号
	<-closeSig
	server.Close()
}

func (gate *WSGate) WriteMsg(msg interface{}) {
	if gate.agent != nil {
		gate.agent.WriteMsg(msg)
	}
}

type WSGateAgent struct {
	network.WSAgent
	gate *WSGate
}
