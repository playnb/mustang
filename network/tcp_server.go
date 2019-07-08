package network

import (
	"net"
	"sync"

	"cell/common/mustang/log"
	"fmt"
)

//TCPServer TCP服务器
type TCPServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	NewAgent        func(*TCPConn) IAgent
	ln              net.Listener
	conns           ConnSet
	mutexConns      sync.Mutex
	wg              sync.WaitGroup
	closeFlag       bool

	// msg parser
	msgParser *MsgParser
}

//Start 启动
func (server *TCPServer) Start() {
	server.init()
	go server.run()
}

func (server *TCPServer) String() string {
	return fmt.Sprintf("TCPServer")
}

//检查参数合法性
func (server *TCPServer) checkValid() {
	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Trace("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 2
		log.Trace("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
}

func (server *TCPServer) init() {
	server.checkValid()

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	} else {
		log.Info("绑定服务器端口成功 @%s", server.Addr)
	}

	server.ln = ln
	server.conns = make(ConnSet)
	server.closeFlag = false
	server.msgParser = NewMsgParser()
}

//绑定端口,等待连接
func (server *TCPServer) run() {
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if server.closeFlag {
				log.Error("%s 已经关闭 Accept Error: %v", server, err)
				return
			}
			log.Error("%s Accept Error: %v", server, err)
			continue
		}

		//加入连接集合
		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Warning("%s: too many connections ...", server)
			continue
		}
		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()

		server.wg.Add(1)

		tcpConn := newTCPConn(conn, server.PendingWriteNum, server.msgParser)
		tcpConn.SetAgentObj(server.NewAgent(tcpConn), server)

		go tcpConn.SendLoop()
		go tcpConn.RecvLoop()
	}
}

//OnTCPConnClose TCPConn关闭时候的回调(在TcpConn.RecvLoop线程中调用)
func (server *TCPServer) OnTCPConnClose(tcpConn *TCPConn) {
	server.mutexConns.Lock()
	delete(server.conns, tcpConn.conn)
	server.mutexConns.Unlock()

	server.wg.Done()
}

func (server *TCPServer) Close() {
	server.closeFlag = true
	server.ln.Close()

	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}
	server.conns = make(ConnSet)
	server.mutexConns.Unlock()

	server.wg.Wait()
}
