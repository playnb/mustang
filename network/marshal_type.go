package network

import "github.com/gogo/protobuf/proto"

type MarshalToProto interface {
	MarshalTo([]byte) (int, error)
	Size() int
	proto.Message
}
