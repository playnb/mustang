package cell_base

import (
	"cell/common/config"
	"cell/common/mustang/log"
	"cell/common/protocol/msg"
	"fmt"
	"math/rand"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var initFunctionList []func(map[uint64]IRpcClientDef)

func RegRpcClientDef(f func(map[uint64]IRpcClientDef)) {
	initFunctionList = append(initFunctionList, f)
}

func InitAll(m map[uint64]IRpcClientDef) {
	for _, f := range initFunctionList {
		f(m)
	}
}

type HasComeinService interface {
	ComeinService(ctx context.Context, in *msg.RPC_ServiceDefine, opts ...grpc.CallOption) (*msg.RPC_ReturnMsg, error)
	CheckAlive(ctx context.Context, in *msg.RPC_ReturnMsg, opts ...grpc.CallOption) (*msg.RPC_ReturnMsg, error)
}

type RpcClient struct {
	Client     HasComeinService
	ClientType uint64
	ClientId   uint64
	ClientAddr string
	ClientConn *grpc.ClientConn
	Define     IRpcClientDef
}

type IRpcClientDef interface {
	IsReady() bool
	Init(ServiceType msg.RPC_ServiceType, DefaultClient HasComeinService, NewFunc func(cc *grpc.ClientConn) HasComeinService)
	GetClient(uint64) HasComeinService
	RandomClient() HasComeinService
	Connect(*msg.IS_RegisterService) *RpcClient
}

////////////////////////////////////////////////////
type RpcClientDef struct {
	ServiceType   msg.RPC_ServiceType
	Enable        bool
	Clients       sync.Map
	DefaultClient HasComeinService
	NewFunc       func(cc *grpc.ClientConn) HasComeinService
}

func (def *RpcClientDef) Init(
	ServiceType msg.RPC_ServiceType,
	DefaultClient HasComeinService,
	NewFunc func(cc *grpc.ClientConn) HasComeinService) {
	def.ServiceType = ServiceType
	def.DefaultClient = DefaultClient
	def.NewFunc = NewFunc
}

func (def *RpcClientDef) Connect(serv *msg.IS_RegisterService) *RpcClient {
	rpcClient := def.MakeRpcClient(serv)
	if rpcClient == nil {
		return nil
	}
	client := def.NewFunc(rpcClient.ClientConn)
	def.Clients.Store(serv.GetServiceId(), client)
	//def.Clients.Store(rpcClient.ClientId, client)

	rpcClient.Client = client
	rpcClient.Define = def
	return rpcClient
}

func (def *RpcClientDef) IsReady() bool {
	if !def.Enable {
		fmt.Printf("%s 没有准备好\n", config.GetServiceName(uint64(def.ServiceType)))
		return false
	}
	return true
}

func (def *RpcClientDef) GetClient(sid uint64) HasComeinService {
	c, ok := def.Clients.Load(sid)
	if ok {
		return c.(HasComeinService)
	}
	return def.DefaultClient
}

func (def *RpcClientDef) RandomClient() HasComeinService {
	clients := make([]HasComeinService, 0, 10)
	def.Clients.Range(func(k, v interface{}) bool {
		clients = append(clients, v.(HasComeinService))
		return true
	})

	if len(clients) > 0 {
		return clients[rand.Int()%len(clients)]
	} else {
		return nil
	}
}

func (def *RpcClientDef) MakeRpcClient(serv *msg.IS_RegisterService) *RpcClient {
	conn, err := grpc.Dial(serv.GetServiceAddr(), grpc.WithInsecure())
	if err != nil {
		log.Error("=============================" + err.Error())
		return nil
	}
	log.Trace("%d 连接服务: ==> %s %s %d",
		def.ServiceType,
		config.GetServiceName(serv.GetServiceType()),
		serv.GetServiceAddr(),
		serv.GetServiceType())

	rpcClient := &RpcClient{}
	rpcClient.ClientType = serv.GetServiceType()
	rpcClient.ClientId = serv.GetServiceId()
	rpcClient.ClientAddr = serv.GetServiceAddr()
	rpcClient.ClientConn = conn
	return rpcClient
}

func (def *RpcClientDef) RandomClientID() uint64 {
	var list []uint64

	def.Clients.Range(func(k, v interface{}) bool {
		list = append(list, k.(uint64))
		return true
	})
	if len(list) > 0 {
		return list[rand.Int()%len(list)]
	}
	return 0
}


////////////////////////////////////////////////////

//==========================================

//==========================================
