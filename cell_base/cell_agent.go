package cell_base

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/network"
	"github.com/playnb/protocol/msg"
)

type CellAgent struct {
	network.TCPAgent
	isValid         bool
	asServer        bool
	serverServiceID uint64
	clientServiceID uint64
	base            *CellBase

	OnCloseFunc func(agent *CellAgent)
}

func NewCellAgent() *CellAgent {
	agent := &CellAgent{}
	agent.isValid = false
	agent.asServer = false
	return agent
}

func (agent *CellAgent) IsValid() bool {
	return agent.isValid
}

func (agent *CellAgent) GetServiceID() uint64 {
	if agent.asServer {
		return agent.clientServiceID
	} else {
		return agent.serverServiceID
	}
}

func (agent *CellAgent) ConnectFunc() {
	log.Trace("==============> 产生新链接 %v(Server:%v) [%d==>%d]", agent, agent.asServer, agent.clientServiceID, agent.serverServiceID)

	if agent.asServer == false {
		agent.base.AddAgent(agent)

		cmd := &msg.MSG_ServerLogin{}
		cmd.ServerId = agent.clientServiceID//proto.Uint64(agent.clientServiceID)
		agent.WriteMsg(cmd)
	}
}

func (agent *CellAgent) CloseFunc() {
	log.Trace("==============> 链接断掉 %v(Server:%v) [%d==>%d]", agent, agent.asServer, agent.clientServiceID, agent.serverServiceID)

	if agent.OnCloseFunc != nil {
		agent.OnCloseFunc(agent)
	}

	agent.base.serverMutex.Lock()
	defer agent.base.serverMutex.Unlock()

	agent.isValid = false
	if agent.asServer == true {
		delete(agent.base.allClients, agent.serverServiceID)
	} else {
		delete(agent.base.allService, agent.clientServiceID)
	}
}
