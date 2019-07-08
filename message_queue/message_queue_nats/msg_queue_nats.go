package message_queue_nats

import (
	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/message_queue"

	"github.com/nats-io/nats"
	"github.com/nats-io/nats/encoders/protobuf"
)

type MsgQueue struct {
	msgQueInited bool // 服务器开启nats标志
	natsConn     *nats.EncodedConn
}

func NewMsgQueue() message_queue.IMsgQueue {
	p := &MsgQueue{}
	p.msgQueInited = false
	p.natsConn = nil
	return p
}

func (self *MsgQueue) SetMsgQueInited() {
	self.msgQueInited = true
}

func (self *MsgQueue) getMsgQueInited() bool {
	return self.msgQueInited
}

func (self *MsgQueue) GetMsgQue() *nats.EncodedConn {
	if self != nil && self.natsConn != nil {
		return self.natsConn
	}

	return nil
}

func (self *MsgQueue) isConnected() bool {
	if self == nil || self.natsConn == nil || self.natsConn.Conn == nil {
		return false
	}

	return self.natsConn.Conn.IsConnected()
}

func (self *MsgQueue) connectGnatsd(serviceName string) bool {
	nats.RegisterEncoder(protobuf.PROTOBUF_ENCODER, &protobuf.ProtobufEncoder{})

	msgQueServerAddr := message_queue.GetMsgQueueServerAddress()
	if len(msgQueServerAddr) == 0 {
		msgQueServerAddr = nats.DefaultURL
		log.Error("%s connect with gnatsd by default server address!", serviceName)
	}

	var err error
	nc, err := nats.Connect(msgQueServerAddr)
	if err != nil {
		log.Error("%s can't connect: %v\n", serviceName, err)
	}

	// nats.JSON_ENCODER
	// protobuf.PROTOBUF_ENCODER
	self.natsConn, err = nats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		log.Error("%s failed to create an encoded connection: %v, serverAddr(%s)", serviceName, err, msgQueServerAddr)
	} else {
		log.Debug("%s connect with gnatsd(%s) successfully!", serviceName, msgQueServerAddr)
	}

	return self.natsConn != nil
}

func (self *MsgQueue) Subscribe(t string, f interface{}) error {
	if self.natsConn == nil {
		return nil
	}

	_, err := self.natsConn.Subscribe(t, f)
	if err != nil {
		log.Error("MsgQueue.Subscribe %s", err.Error())
	}
	return err
}

func (self *MsgQueue) Publish(topic string, msgData interface{}) {
	if self.natsConn == nil {
		return
	}

	err := self.natsConn.Publish(topic, msgData)
	if err != nil {
		log.Error("Publish err: %v", err)
	}

	err = self.natsConn.Flush()
	if err != nil {
		log.Error("Flush err: %v", err)
	}
}

func (self *MsgQueue) InitMsgQue(serviceName string) bool {
	if self.getMsgQueInited() == true && (self.GetMsgQue() == nil || self.isConnected() == false) {
		return self.connectGnatsd(serviceName)
	}

	return false
}
