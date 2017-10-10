package rpc

import (
	proto "github.com/gogo/protobuf/proto"
)

type CallUnit struct {
	ServiceID  uint32
	ClienID    uint32
	CallBack   func(proto.Message)
	Args       proto.Message
	ClientData interface{}
}

func Call(service_id uint32, rpc_name string, args proto.Message, client_data interface{}, call_back func(ret proto.Message)) {

}
