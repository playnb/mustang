package network

import (
	"github.com/playnb/mustang/log"
	"net"
	"sync"
	"time"
)

type TCPClient struct {
	sync.Mutex
	Addr            string
	ConnectInterval time.Duration
	PendingWriteNum int
	MaxTry          int
	NewAgent        func(*TCPConn) IAgent
	tcpConn         *TCPConn
	closeFlag       bool
	agent           IAgent

	// msg parser
	msgParser *MsgParser
}

func (client *TCPClient) Start() {
	client.init()
	go client.connect()
}

func (client *TCPClient) check_valid() {
	if client.ConnectInterval <= 0 {
		client.ConnectInterval = 3 * time.Second
		log.Trace("invalid ConnectInterval, reset to %v", client.ConnectInterval)
	}
	if client.PendingWriteNum <= 0 {
		client.PendingWriteNum = 100
		log.Trace("invalid PendingWriteNum, reset to %v", client.PendingWriteNum)
	}
	if client.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
}

func (client *TCPClient) init() {
	client.Lock()
	defer client.Unlock()

	client.check_valid()

	client.closeFlag = false
	client.msgParser = NewMsgParser()
}

//创建net.Conn连接
func (client *TCPClient) dial() net.Conn {
	try := 0
	log.Trace("尝试创建连接  %v", client.Addr)
	for {
		conn, err := net.Dial("tcp", client.Addr)
		if err == nil {
			return conn
		}
		log.Trace("连接失败 %v error: %v try: %d", client.Addr, err, try)

		try = try + 1
		if try > client.MaxTry {
			return nil
		}

		time.Sleep(client.ConnectInterval)
		log.Trace("尝试重连 %v try: %d", client.Addr, try)
		continue
	}
}

func (client *TCPClient) connect() {
	if client.tcpConn != nil {
		return
	}

	conn := client.dial()
	if conn == nil {
		log.Error("创建连接失败  %v", client.Addr)
		return
	}
	log.Trace("创建连接成功  %v", client.Addr)

	client.tcpConn = newTCPConn(conn, client.PendingWriteNum, client.msgParser)
	client.agent = client.NewAgent(client.tcpConn)
	client.agent.Run()

	client.tcpConn.Close()
	client.agent.OnClose()
	client.tcpConn = nil
	client.agent = nil
}

func (client *TCPClient) Close() {
	client.Lock()
	if client.closeFlag {
		return
	}
	client.closeFlag = true
	if client.tcpConn != nil {
		client.tcpConn.Close()
	}
	client.Unlock()
}

func (client *TCPClient) WriteMsg(msg interface{}) {
	if client.agent != nil {
		client.agent.WriteMsg(msg)
	}
}
