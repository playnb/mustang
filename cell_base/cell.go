package cell_base

import (
	"github.com/playnb/config"
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/network"
	"github.com/playnb/mustang/network/msg_processor"
	"github.com/playnb/protocol"
	"github.com/playnb/protocol/msg"
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/playnb/mustang/util"

	"github.com/gogo/protobuf/proto"
	"qiniupkg.com/x/errors.v7"
)

type updataFunction func(*msg.IS_RequireAllService_Reply)

type tMsgData struct {
	MsgData  proto.Message
	MsgIndex uint16
	Agent    network.IAgent
}

type CellBase struct {
	tcpService         *network.TCPServer
	serverProcessor    network.IMsgProcessor //处理其他Cell来的消息
	clientProcessor    network.IMsgProcessor //处理其他Cell来的消息
	indexServiceClient msg.IndexServiceClient
	allClients         map[uint64]*CellAgent //所有的客户端链接(主动链接 asServer==true)
	allService         map[uint64]*CellAgent //所有的服务器链接(被动链接 asServer==false)

	serverListVersion uint64
	updateServerList  []updataFunction
	CellType          uint64
	Name              string
	ID                uint64
	ServerAddr        string

	serverMutex sync.Mutex
}

var _CellBaseOnce sync.Once
var _instance *CellBase = nil

func GetCellBase() *CellBase {
	_CellBaseOnce.Do(func() {
		_instance = NewCellBase()
	})
	return _instance
}

func NewCellBase() *CellBase {
	ins := &CellBase{}
	ins.serverProcessor = msg_processor.NewPbProcessor()
	ins.clientProcessor = msg_processor.NewPbProcessor()
	ins.allClients = make(map[uint64]*CellAgent)
	ins.allService = make(map[uint64]*CellAgent)
	ins.updateServerList = make([]updataFunction, 0, 1)
	protocol.KnowAllMsg(ins)
	return ins
}

func (cell *CellBase) String() string {
	if cell == nil {
		return ""
	}

	return fmt.Sprintf("Name:%s, ID:%d, CellType:%d, ServerAddr:%d", cell.Name, cell.ID, cell.CellType, cell.ServerAddr)
}

//计算服务的唯一ID
func (cell *CellBase) MakeServiceID(sid uint64) uint64 {
	return (cell.ID * 100) + sid
}

//解析Cell的ID
func (cell *CellBase) ParseCellID(sid uint64) uint64 {
	return sid / 100
}

func (cell *CellBase) AddUpdateServerList(foo updataFunction) {
	_instance.updateServerList = append(_instance.updateServerList, foo)
}

//初始化Cell，连接IndexService
func (cell *CellBase) Init(indexAddr string) {
	if true {
		//创建到IndexService的RPC连接
		go func() {
			for true {
				conn, err := grpc.Dial(indexAddr, grpc.WithInsecure())
				if err != nil {
					cell.indexServiceClient = nil
					log.Error("=============================创建Index连接失败@" + err.Error())
				} else {
					cell.indexServiceClient = msg.NewIndexServiceClient(conn)
					log.Trace("=============================创建Index连接成功@" + indexAddr)
					_, err := cell.indexServiceClient.LoginIndexService(context.Background(),
						&msg.IS_RequireLogin{
							ServiceId:   cell.ID,         //proto.Uint64(cell.ID),
							ServiceAddr: cell.ServerAddr, //proto.String(cell.ServerAddr),
							ServiceType: 0,               //proto.Uint64(0),
							TcpAddr:     cell.ServerAddr, //proto.String(cell.ServerAddr),
						})
					if err != nil {
						cell.indexServiceClient = nil
						log.Error("=============================创建Index连接失败2@" + err.Error())
					} else {
						break
					}
				}
				time.Sleep(time.Second)
			}
		}()

		go func() {
			t10s := time.NewTicker(time.Second * 10)
			t1s := time.NewTicker(time.Second)
			for {
				select {
				case <-t10s.C:
					if cell.indexServiceClient != nil {
						//						log.Trace("====================> RequireAllService")
						reply, err := cell.indexServiceClient.RequireAllService(context.Background(), &msg.IS_RequireAllService{})
						if err != nil {
							log.Error("[CellBase] MainLoop %s", err.Error())
						} else {

							//for _, proxy := range reply.GetServiceList() {
							//	log.Trace("[CellBase] MainLoop %v", proxy)
							//}

							if cell.updateServerList != nil {
								for _, foo := range cell.updateServerList {
									foo(reply)
								}
							}
						}
					}
				case <-t1s.C:
					if cell.indexServiceClient != nil {
						//log.Trace("====================> MaintainAlive")
						reply, _ := cell.indexServiceClient.MaintainAlive(util.StandardTimeOutContext(), &msg.IS_MaintainAlive{
							ServiceId: cell.ID, //proto.Uint64(cell.ID),
						})
						if reply != nil {
							cell.serverListVersion = reply.GetServiceListVersion()
						}
					}
				}
			}
		}()
	}
}

func (cell *CellBase) InitTcpService(serviceID uint64, serverAddr string) {
	cell.ServerAddr = serverAddr
	cell.tcpService = &network.TCPServer{}
	cell.tcpService.Addr = cell.ServerAddr
	cell.tcpService.MaxConnNum = 50000
	cell.tcpService.PendingWriteNum = 1024
	if config.DebugMode {
		cell.tcpService.PendingWriteNum = 1
	}
	cell.tcpService.NewAgent = func(conn *network.TCPConn) network.IAgent {
		ag := NewCellAgent()
		ag.SetConn(conn)
		ag.SetName(cell.Name)
		ag.SetMsgProcessor(cell.clientProcessor)
		ag.asServer = true
		ag.serverServiceID = serviceID
		ag.clientServiceID = 0
		ag.base = cell
		//ag.SetUserData(NewNetUser(ag))
		return ag
	}
	cell.tcpService.Start()
}

func (cell *CellBase) ConnectToServer(clientID uint64, serverID uint64, addr string, closeFunc func(*CellAgent)) {
	client := &network.TCPClient{}
	client.MaxTry = 3
	client.ConnectInterval = 3 * time.Second
	client.PendingWriteNum = 4096
	client.Addr = addr
	client.NewAgent = func(conn *network.TCPConn) network.IAgent {
		ag := NewCellAgent()
		ag.SetConn(conn)
		ag.SetName(cell.Name)
		ag.SetMsgProcessor(cell.serverProcessor)
		ag.asServer = false
		ag.clientServiceID = clientID
		ag.serverServiceID = serverID
		ag.base = cell
		ag.OnCloseFunc = closeFunc
		ag.TcpMsgChanCount = 0
		return ag
	}
	client.Start()
	log.Trace("%s 尝试连接服务 %s", cell.String(), addr)
}

func (cell *CellBase) SendCmdToClient(id uint64, data network.MarshalToProto) bool {
	if id == 0 {
		return false
	}

	cell.serverMutex.Lock()
	defer cell.serverMutex.Unlock()

	if agent, ok := cell.allClients[id]; ok {
		agent.WriteMsg(data)
		return true
	}
	return false
}

func (cell *CellBase) SendCmdToServer(id uint64, data network.MarshalToProto) bool {
	if id == 0 {
		return false
	}

	cell.serverMutex.Lock()
	defer cell.serverMutex.Unlock()

	if agent, ok := cell.allService[id]; ok {
		agent.WriteMsg(data)
		return true
	}
	return false
}

func (cell *CellBase) ForwardToUser(uid uint64, sid uint64, data network.MarshalToProto) bool {
	if sid == 0 {
		return false
	}

	cell.serverMutex.Lock()
	defer cell.serverMutex.Unlock()

	if agent, ok := cell.allClients[sid]; ok {
		forward := PackForwardToUser(uid, nil, data, agent.MsgProcessor())
		if forward == nil {
			return false
		}
		agent.WriteMsg(forward)

		return true
	}
	return false
}

func (cell *CellBase) ForwardToUsers(uids []uint64, sid uint64, data network.MarshalToProto) bool {
	if sid == 0 {
		return false
	}

	cell.serverMutex.Lock()
	defer cell.serverMutex.Unlock()

	if agent, ok := cell.allClients[sid]; ok {
		forward := PackForwardToUser(0, uids, data, agent.MsgProcessor())
		if forward == nil {
			return false
		}
		agent.WriteMsg(forward)

		return true
	}
	return false
}

func (cell *CellBase) HasServer(id uint64) bool {
	cell.serverMutex.Lock()
	defer cell.serverMutex.Unlock()

	if _, ok := cell.allService[id]; ok {
		return true
	}
	return false
}

func (cell *CellBase) HasClient(id uint64) bool {
	cell.serverMutex.Lock()
	defer cell.serverMutex.Unlock()

	if _, ok := cell.allClients[id]; ok {
		return true
	}
	return false
}

func (cell *CellBase) RegisterService(servID uint64, servName string, servType uint64, servAddr string) (*msg.IS_RegisterService_Reply, error) {
	if cell.indexServiceClient != nil {
		return cell.indexServiceClient.RegisterService(context.Background(), &msg.IS_RegisterService{
			ServiceId:   servID,   //proto.Uint64(servID),
			ServiceAddr: servAddr, //proto.String(servAddr),
			ServiceType: servType, //proto.Uint64(servType),
			TcpAddr:     servAddr, //proto.String(servAddr),
		})
	}
	return nil, errors.New("Index连接不存在")
}

func (cell *CellBase) KnowMsg(sample interface{}) {
	cell.RegisterServerMsg(sample, nil)
	cell.RegisterClientMsg(sample, nil)
	cell.RegisterClientMsg(&msg.MSG_ServerLogin{}, on_MSG_ServerLogin)
}

func (cell *CellBase) RegisterServerMsg(sample interface{}, msgHandler network.MsgHandler) {
	cell.serverProcessor.Register(sample, msgHandler)
}

func (cell *CellBase) RegisterClientMsg(sample interface{}, msgHandler network.MsgHandler) {
	cell.clientProcessor.Register(sample, msgHandler)
}

func (cell *CellBase) AddAgent(agent *CellAgent) {
	cell.serverMutex.Lock()
	defer cell.serverMutex.Unlock()

	agent.isValid = true
	if agent.asServer == true {
		cell.allClients[agent.clientServiceID] = agent
	} else {
		cell.allService[agent.serverServiceID] = agent
	}
}

func on_MSG_ServerLogin(agent network.SampleAgent, msgData interface{}, msgIndex uint16, data interface{}) {
	cmd := msgData.(*msg.MSG_ServerLogin)
	ca := agent.(*CellAgent)
	if ca.asServer {
		ca.clientServiceID = cmd.GetServerId()
	} else {
		ca.serverServiceID = cmd.GetServerId()
	}
	GetCellBase().AddAgent(ca)
	log.Trace("====>MSG_ServerLogin: %v", cmd)
}
