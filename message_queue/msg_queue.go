package message_queue

import (
	nats "github.com/nats-io/nats.go"
)

type IMsgQueue interface {
	SetMsgQueInited()
	Subscribe(t string, f interface{}) error
	Publish(topic string, msgData interface{})
	InitMsgQue(serviceName string) bool
}

var msgQueueServerAddr = nats.DefaultURL

func InitMsgQueueServerAddress(addr string) {
	if len(addr) == 0 {
		return
	}

	msgQueueServerAddr = addr
}

func GetMsgQueueServerAddress() string {
	return msgQueueServerAddr
}
