package processor

import (
	"github.com/playnb/mustang/network"
	"github.com/playnb/mustang/util"
	"encoding/binary"
	"sync"
	"sync/atomic"
)

/*
| 2B Version| 4B MessageType| 4B MessageID| MessageName(1B size+data)| 4B ErrorCode| Data
*/

const (
	MessageTypeRpcMask     = uint32(7)
	MessageTypeRpc         = uint32(4)
	MessageTypeRpcRequest  = uint32(4 + 1)
	MessageTypeRpcResponse = uint32(4 + 2)
)

type ICanPack interface {
	PackMsg() *MsgDef
	MsgName() string
}

type HandleFunc func(agent network.SampleAgent, md *MsgDef, clientData interface{}) bool

type MsgDef struct {
	Version     uint16
	MessageType uint32
	MessageID   uint32
	MessageName string
	ErrorCode   uint32
	Data        []byte

	SourceID uint64
}

func (md *MsgDef) IsNotifyMsg() bool {
	return (md.MessageType & MessageTypeRpc) == 0
}
func (md *MsgDef) IsRpcRequest() bool {
	return (md.MessageType & MessageTypeRpcRequest) == MessageTypeRpcRequest
}
func (md *MsgDef) IsRpcResponse() bool {
	return (md.MessageType & MessageTypeRpcResponse) == MessageTypeRpcResponse
}
func (md *MsgDef) SetRpcRequest() {
	md.MessageType = md.MessageType&(^MessageTypeRpcMask) | MessageTypeRpcRequest
}
func (md *MsgDef) SetRpcResponse() {
	md.MessageType = md.MessageType&(^MessageTypeRpcMask) | MessageTypeRpcResponse
}

func (md *MsgDef) HeaderSize() int {
	return 2 + 4 + 4 + 1 + len(md.MessageName) + 4
}

func (md *MsgDef) UnpackMsg(data []byte) {
	index := 0

	md.Version = binary.BigEndian.Uint16(data[index:])
	index += 2

	md.MessageType = binary.BigEndian.Uint32(data[index:])
	index += 4

	md.MessageID = binary.BigEndian.Uint32(data[index:])
	index += 4

	nameSize := int(data[index])
	index += 1

	md.MessageName = string(data[index : index+nameSize])
	index += nameSize

	md.ErrorCode = binary.BigEndian.Uint32(data[index:])
	index += 4

	md.Data = data[index:]
}

func (md *MsgDef) PackMsg(buf *util.BuffData) *util.BuffData {
	size := md.HeaderSize()
	if buf == nil {
		buf = util.MakeBuffData(size + len(md.Data))
	} else {
		buf.ChangeIndex(-size)
	}
	data := buf.GetPayload()
	index := 0

	binary.BigEndian.PutUint16(data[index:], md.Version)
	index += 2

	binary.BigEndian.PutUint32(data[index:], md.MessageType)
	index += 4

	binary.BigEndian.PutUint32(data[index:], md.MessageID)
	index += 4

	data[index] = byte(len(md.MessageName))
	index += 1

	copy(data[index:], []byte(md.MessageName))
	index += len(md.MessageName)

	binary.BigEndian.PutUint32(data[index:], md.ErrorCode)
	index += 4

	if md.Data != nil {
		copy(data[index:], md.Data)
	}

	return buf
}

//////////////////////////
type CallBase struct {
	agent         network.IAgent
	_callSequence uint32
	_callSub      map[uint32]func(*MsgDef)
	_callMutex    sync.Mutex
}

func (c *CallBase) Agent() network.IAgent {
	return c.agent
}

func (c *CallBase) GetCallSequence() uint32 {
	return atomic.AddUint32(&c._callSequence, 1)%40000000 + 1000
}

func (c *CallBase) AddCallback(id uint32, f func(*MsgDef)) {
	c._callMutex.Lock()
	defer c._callMutex.Unlock()
	c._callSub[id] = f
}

func (c *CallBase) RunCallback(md *MsgDef) {
	c._callMutex.Lock()
	defer c._callMutex.Unlock()
	if f, ok := c._callSub[md.MessageID]; ok {
		f(md)
		delete(c._callSub, md.MessageID)
	}
}

func (c *CallBase) Init(agent network.IAgent) {
	c._callSub = make(map[uint32]func(*MsgDef))
	c._callSequence = 0
	c.agent = agent
}

//////////////////////////
type NotifyBase struct {
	agent network.SampleAgent
}

func (mb *NotifyBase) SetAgent(agent network.SampleAgent) {
	mb.agent = agent
}

func (mb *NotifyBase) Agent() network.SampleAgent {
	return mb.agent
}
