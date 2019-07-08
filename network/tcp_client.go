package network

import (
	"net"
	"sync"
	"time"

	"cell/common/mustang/log"
)

//TCPClient Tcp客户端
type TCPClient struct {
	sync.Mutex
	Addr             string
	ConnectInterval  time.Duration
	PendingWriteNum  int
	MaxTry           int
	NewAgent         func(*TCPConn) IAgent
	ConnectTryFailed func()
	tcpConn          *TCPConn
	closeFlag        bool
	agent            IAgent
	name             string

	// msg parser
	PacketParser *MsgParser
}

//Start 启动
func (client *TCPClient) Start() {
	client.init()
	go client.connect()
}

func (client *TCPClient) checkValid() {
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
	client.checkValid()
	client.closeFlag = false
	if client.PacketParser == nil {
		client.PacketParser = NewMsgParser()
	}
}

//创建net.Conn连接
func (client *TCPClient) dial() net.Conn {
	try := 0
	log.Trace("[%s] 尝试创建连接  %v", client.String(), client.Addr)
	for {
		conn, err := net.Dial("tcp", client.Addr)
		if err == nil {
			return conn
		}
		log.Trace("[%s] 连接失败 %v error: %v try: %d", client.String(), client.Addr, err, try)

		try = try + 1
		if try > client.MaxTry {
			if client.ConnectTryFailed != nil {
				client.ConnectTryFailed()
			}
			return nil
		}

		time.Sleep(client.ConnectInterval)
		log.Trace("[%s] 尝试重连 %v try: %d", client, client.Addr, try)
		continue
	}
}

func (client *TCPClient) connect() {
	defer log.PrintPanicStack()
	if client.tcpConn != nil {
		return
	}

	conn := client.dial()
	if conn == nil {
		log.Error("[%s] 创建连接失败  %v", client, client.Addr)
		return
	}
	log.Trace("[%s] 创建连接成功  %v", client, client.Addr)

	client.tcpConn = newTCPConn(conn, client.PendingWriteNum, client.PacketParser)
	client.agent = client.NewAgent(client.tcpConn)
	client.tcpConn.SetAgentObj(client.agent, client)

	go client.tcpConn.SendLoop()
	go client.tcpConn.RecvLoop()
}

//OnTCPConnClose TCPConn关闭时候的回调(在TcpConn.RecvLoop线程中调用)
func (client *TCPClient) OnTCPConnClose(tcpConn *TCPConn) {
	client.tcpConn = nil
}

func (client *TCPClient) Close() {
	client.Lock()
	defer client.Unlock()
	if client.closeFlag {
		return
	}
	client.closeFlag = true
	if client.tcpConn != nil {
		client.tcpConn.Terminate()
	}
}

func (client *TCPClient) WriteMsg(msg MarshalToProto) {
	if client.agent != nil {
		client.agent.WriteMsg(msg)
	}
}

func (client *TCPClient) String() string {
	return client.name
}
func (client *TCPClient) SetName(name string) {
	client.name = name
}
