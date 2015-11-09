package protobuf

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/network"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	"reflect"
)

//-----------------------------------------------
//| 16bit NameSize | MsgName | ptotobuf message |
//-----------------------------------------------

type MsgInfo struct {
	msgType    reflect.Type
	msgName    string
	msgHandler network.MsgHandler
}

type Processor struct {
	littleEndian bool
	msgInfo      map[string]*MsgInfo
}

func NewProcessor() *Processor {
	p := new(Processor)
	p.littleEndian = false
	p.msgInfo = make(map[string]*MsgInfo)
	return p
}

func (p *Processor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

func (p *Processor) Register(msg proto.Message, msgHandler network.MsgHandler) {
	msgType := reflect.TypeOf(msg)
	msgName := msgType.String()[1:]
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("protobuf message pointer required")
	}
	if len(p.msgInfo) >= math.MaxUint16 {
		log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
	}

	i := new(MsgInfo)
	i.msgType = msgType
	i.msgName = msgName
	i.msgHandler = msgHandler
	p.msgInfo[msgName] = i
}

/*
func (p *Processor) SetHandler(msg proto.Message, msgHandler MsgHandler) {
	msgType := reflect.TypeOf(msg)
	msgName := msgType.String()[1:]
	i, ok := p.msgInfo[msgName]
	if !ok {
		log.Fatal("message %s not registered", msgType)
	}

	i.msgHandler = msgHandler
}
*/

func (p *Processor) Handler(reciver network.IAgent, data []byte, clientData interface{}) (bool, error) {
	if len(data) < 2 {
		return false, errors.New("protobuf data too short")
	}

	nameLen := int(data[0])*265 + int(data[1])
	if len(data) < 2+nameLen {
		return false, errors.New("protobuf data too short")
	}

	msgName := string(data[2 : nameLen+2])
	i, ok := p.msgInfo[msgName]
	if !ok {
		s := fmt.Sprintf("message %s not registered", msgName)
		log.Fatal(s)
		return false, errors.New(s)
	}
	if i.msgHandler == nil {
		s := fmt.Sprintf("message %s not handler", msgName)
		log.Fatal(s)
		return false, errors.New(s)
	}

	msg := reflect.New(i.msgType.Elem()).Interface().(proto.Message)
	if e := proto.Unmarshal(data[nameLen+2:], msg); e != nil {
		return false, e
	}

	if i.msgHandler != nil {
		i.msgHandler(reciver, i.msgType.Elem(), msg, []interface{}{data, clientData})
	}
	return true, nil
}

func (p *Processor) PackMsg(msg proto.Message) ([]byte, error) {
	msgType := reflect.TypeOf(msg)
	msgName := msgType.String()[1:]
	nameSize := len(msgName)
	_data, _err := proto.Marshal(msg)
	if _err != nil {
		return nil, _err
	}

	data := make([]byte, 2+nameSize+len(_data))
	data[0] = byte(int(nameSize) / 256)
	data[1] = byte(int(nameSize) % 256)
	copy(data[2:nameSize+2], []byte(msgName))
	copy(data[nameSize+2:], []byte(_data))
	return data, nil
}
