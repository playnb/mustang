package cell_base

import (
	"cell/common/mustang/log"
	"cell/common/protocol/msg"
	"sync"

	"golang.org/x/net/context"
)

type CanCheckAlive struct {
}

func (c *CanCheckAlive) CheckAlive(ctx context.Context, in *msg.RPC_ReturnMsg) (*msg.RPC_ReturnMsg, error) {
	return &msg.RPC_ReturnMsg{}, nil
}

type IRpcService interface {
	ID() uint64
	Type() uint64
	Addr() string
	Name() string

	OnConnectServiceOK(*msg.IS_RegisterService)
	OnConnectServiceLost(uint64)
}

type RpcClientMgr struct {
	clients      map[uint64]*RpcClient
	clientsMutex sync.RWMutex

	allRpcServiceDef map[uint64]IRpcClientDef
	hostService      IRpcService
	roomClients      map[uint64]*RpcClient
	needService      uint64
	allReady         bool
}

func (mgr *RpcClientMgr) CheckNeedService(sType uint64) bool {
	return (mgr.needService & uint64(1<<sType)) != 0
}

func (mgr *RpcClientMgr) AllReady() bool {
	if mgr.allReady == false {
		for k, v := range mgr.allRpcServiceDef {
			if mgr.CheckNeedService(k) && v.IsReady() == false {
				return false
			}
		}
	}
	mgr.allReady = true
	return mgr.allReady
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (mgr *RpcClientMgr) NeedService(st msg.RPC_ServiceType) {
	mgr.needService |= uint64(1 << uint64(st))
}

func (mgr *RpcClientMgr) GetRandRoomId() uint64 {
	return 0
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (mgr *RpcClientMgr) GetRpcClient(st msg.RPC_ServiceType, sid uint64) interface{} {
	if v, ok := mgr.allRpcServiceDef[uint64(st)]; ok {
		return v.GetClient(sid)
	}
	return nil
}

func (mgr *RpcClientMgr) GetRpcServiceDef(st msg.RPC_ServiceType) IRpcClientDef {
	if v, ok := mgr.allRpcServiceDef[uint64(st)]; ok {
		return v
	}
	return nil
}

func (mgr *RpcClientMgr) GetIndexClient() msg.IndexServiceClient {
	return mgr.GetRpcClient(msg.RPC_ServiceType_INDEX, 0).(msg.IndexServiceClient)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (mgr *RpcClientMgr) InitRpcClientMgr(rpcService IRpcService) {
	mgr.clients = make(map[uint64]*RpcClient)
	mgr.roomClients = make(map[uint64]*RpcClient)
	mgr.hostService = rpcService

	if len(rpcService.Addr()) > 1 {
		_, err := GetCellBase().RegisterService(rpcService.ID(), rpcService.Name(), rpcService.Type(), rpcService.Addr())

		if err != nil {
			log.Error(err.Error())
		}
	}

	GetCellBase().AddUpdateServerList(func(reply *msg.IS_RequireAllService_Reply) {
		//log.Debug(rpcService.Name())
		for _, serv := range reply.GetServiceList() {
			mgr.ConnectRpcService(serv)
		}
	})

	mgr.allRpcServiceDef = make(map[uint64]IRpcClientDef)
	InitAll(mgr.allRpcServiceDef)
}

func (mgr *RpcClientMgr) ConnectRpcService(serv *msg.IS_RegisterService) {
	if mgr.CheckNeedService(serv.GetServiceType()) == false {
		return
	}
	if serv.GetServiceType() > uint64(msg.RPC_ServiceType_NotMgr) {
		return
	}
	mgr.clientsMutex.RLock()
	_, ok := mgr.clients[serv.GetServiceId()]
	mgr.clientsMutex.RUnlock()

	if ok == false {
		serviceDef, ok := mgr.allRpcServiceDef[serv.GetServiceType()]
		if ok == false {
			return
		}

		rpcClient := serviceDef.Connect(serv)

		if rpcClient != nil && rpcClient.Client != nil {
			mgr.clientsMutex.Lock()
			mgr.clients[serv.GetServiceId()] = rpcClient
			mgr.clientsMutex.Unlock()

			_, err := rpcClient.Client.ComeinService(context.Background(), &msg.RPC_ServiceDefine{
				Id:   mgr.hostService.ID(),   //proto.Uint64(mgr.hostService.ID()),
				Name: mgr.hostService.Name(), //proto.String(mgr.hostService.Name()),
				Type: mgr.hostService.Type(), //proto.Uint64(mgr.hostService.Type()),
				Addr: mgr.hostService.Addr(), //proto.String(mgr.hostService.Addr()),
			})
			if err == nil {
				mgr.hostService.OnConnectServiceOK(serv)
			}
		}
	}
}
