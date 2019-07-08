package cell_base

import "cell/common/mustang/util"

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
