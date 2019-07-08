package processor

import (
	"cell/common/mustang/log"
	"cell/common/mustang/network"
	"cell/common/mustang/util"
	"errors"
)

var NoHandleError = errors.New("没有处理函数的情况")
var NoRegError = errors.New("消息沒有注冊")

type Processor struct {
	handles map[string]HandleFunc
}

func (mp *Processor) Init() {
	mp.handles = make(map[string]HandleFunc)
}

func (mp *Processor) Handler(receiver network.SampleAgent, buffData *util.BuffData, clientData interface{}) (*MsgDef, error) {
	data := buffData.GetPayload()

	md := &MsgDef{}
	md.UnpackMsg(data)

	//log.Debug("收到消息：%v", md)

	return mp.Process(receiver, md, clientData)
}

func (mp *Processor) Process(receiver network.SampleAgent, md *MsgDef, clientData interface{}) (*MsgDef, error) {
	if f, ok := mp.handles[md.MessageName]; ok {
		if f(receiver, md, clientData) {
			return md, nil
		}
		return md, NoHandleError
	} else {
		//没有处理函数的情况
		log.Debug("没有处理函数的情况：%v", md)
		return md, NoRegError
	}
}

func (mp *Processor) ProcessMsg(receiver network.SampleAgent, msg ICanPack) {
	if f, ok := mp.handles[msg.MsgName()]; ok {
		_ = f
	} else {
		//没有处理函数的情况
	}
}

func (mp *Processor) PackMsg(network.MarshalToProto) (*util.BuffData, error) {
	return nil, errors.New("没有实现的函数")
}

func (mp *Processor) Register(interface{}, network.MsgHandler) {
}

func (mp *Processor) RegHandle(name string, f HandleFunc) {
	mp.handles[name] = f
}

////////////////////////////////////////////////////////////////////////////////
