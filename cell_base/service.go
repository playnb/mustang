package cell_base

import "github.com/playnb/mustang/util"

type IService interface {
	util.IService
	RpcClientMgr() *RpcClientMgr
}

type Service struct {
	util.Service
	rpcClientMgr RpcClientMgr
}

func (s *Service) RpcClientMgr() *RpcClientMgr {
	return &s.rpcClientMgr
}
