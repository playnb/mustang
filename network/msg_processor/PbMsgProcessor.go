package msg_processor

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/network"
	"github.com/playnb/mustang/util"
	"github.com/playnb/protocol"

	"github.com/gogo/protobuf/proto"
)

//-----------------------------------------------
//| 16bit MsgIndex |       ptotobuf message     |
//-----------------------------------------------

func NewPbProcessor() network.IMsgProcessor {
	p := &Processor{}
	p.littleEndian = false
	p.msgInfo = make(map[uint16]*MsgInfo)
	return p
}

type MsgInfo struct {
	msgType    reflect.Type
	msgName    string
	msgIndex   uint16
	msgHandler network.MsgHandler
}

func (info *MsgInfo) String() string {
	return fmt.Sprintf("msgType:%v, msgName:%v, msgIndex:%v, msgHandler:%v", info.msgType, info.msgName, info.msgIndex, info.msgHandler)
}

type Processor struct {
	littleEndian      bool
	msgInfo           map[uint16]*MsgInfo
	defaultMsgHandler network.MsgHandler
}

func (p *Processor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

func (p *Processor) Register(sample interface{}, msgHandler network.MsgHandler) {
	if sample == nil {
		p.defaultMsgHandler = msgHandler
		log.Trace("Register Default Msg Handle")
	} else {
		msg := sample.(proto.Message)
		msgType := reflect.TypeOf(msg)
		msgName := msgType.String()[1:]
		msgIndex := protocol.GetMsgIndex(msgName)

		if msgType == nil || msgType.Kind() != reflect.Ptr {
			log.Fatal("protobuf message pointer required")
		}
		if len(p.msgInfo) >= math.MaxUint16 {
			log.Fatal("too many protobuf messages (max = %v)", math.MaxUint16)
		}
		if msgIndex == 0 {
			log.Fatal("%v 消息找不到Index...", msgName)
		}

		i := new(MsgInfo)
		i.msgType = msgType
		i.msgName = msgName
		i.msgIndex = msgIndex
		i.msgHandler = msgHandler
		p.msgInfo[i.msgIndex] = i
		//log.Trace("Register Msg Handle: %s Index: %d", i.msgName, i.msgIndex)
	}
}

//Handler 调用消息映射的处理函数
func (p *Processor) Handler(reciver network.SampleAgent, buffData *util.BuffData, clientData interface{}) (bool, error) {
	data := buffData.GetPayload()
	defer buffData.Release()
	if len(data) < 2 {
		log.Ltp("%v", data)
		return false, errors.New("protobuf data too short")
	}

	msgIndex := uint16(data[0])*256 + uint16(data[1])

	i, ok := p.msgInfo[msgIndex]

	if !ok {
		s := fmt.Sprintf("message %d not registered", msgIndex)
		log.Warning(s)
		return false, errors.New(s)
	}

	//__ft := util.NewFunctionTime("Handler:"+i.msgName, 5)
	//defer __ft.End()

	//TODO  要不直接在MsgInfo里面存一个proto对象 每次Reset/Unmarshal
	msg := reflect.New(i.msgType.Elem()).Interface().(proto.Message)
	if e := proto.Unmarshal(data[2:], msg); e != nil {
		log.Trace("%d %d. %s %v", uint16(data[0]), uint16(data[1]), reciver, e)
		return false, e
	}
	
	//log.Ltp("====== Handler msgInfo ----> index:%d, name:%s msg:%v", i.msgIndex, i.msgName, msg)

	if i.msgHandler != nil {
		i.msgHandler(reciver, msg, i.msgIndex, clientData)
	} else if p.defaultMsgHandler != nil {
		p.defaultMsgHandler(reciver, msg, i.msgIndex, clientData)
	}

	return true, nil
}

func (p *Processor) Handler2(reciver network.IAgent, msgIndex uint16, msgData interface{}) (bool, error) {
	i, ok := p.msgInfo[msgIndex]

	if !ok {
		s := fmt.Sprintf("message %d not registered", msgIndex)
		log.Warning(s)
		log.Dev("%v", p.msgInfo)
		return false, errors.New(s)
	}

	//log.Dev("收到消息 %d %s", msgIndex, i.msgName)

	if i.msgHandler != nil {
		i.msgHandler(reciver, msgData, i.msgIndex, nil)
	}

	return true, nil
}

//打包消息
func (p *Processor) PackMsg(msg network.MarshalToProto) (*util.BuffData, error) {
	msgType := reflect.TypeOf(msg)
	msgName := msgType.String()[1:]
	msgIndex := protocol.GetMsgIndex(msgName)
	if msgIndex == 0 {
		s := fmt.Sprintf("message %s not registered", msgName)
		log.Warning(s)
	}

	{
		m := msg.(network.MarshalToProto)

		//buff := util.MakeBuffData(m.Size() + network.BufHeadOffset + 2)
		buff := util.DefaultPool().Get(m.Size() + network.BufHeadOffset + 2)
		buff.ChangeIndex(network.BufHeadOffset + 2)

		size, err := m.MarshalTo(buff.GetPayload())
		buff.SetSize(size)
		buff.ChangeIndex(-2)
		if err != nil {
			return nil, err
		}
		buff.GetPayload()[0] = byte(int(msgIndex) / 256)
		buff.GetPayload()[1] = byte(int(msgIndex) % 256)
		return buff, nil
	}

	/* {
		_data, _err := proto.Marshal(msg.(proto.Message))
		if _err != nil {
			return nil, _err
		}

		//log.Ltp("====== PackMsg msgInfo ----> index:%d, name:%s, msg:%v", msgIndex, msgName, msg)

		data := make([]byte, 2+len(_data))
		data[0] = byte(int(msgIndex) / 256)
		data[1] = byte(int(msgIndex) % 256)
		copy(data[2:], []byte(_data))
		return data, nil
	}*/
}
