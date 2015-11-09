package network

import (
	"github.com/playnb/mustang/log"
	"net"
	"sync"
)

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

func (server *TCPServer) Start() {
	server.init()
	go server.run()
}

//检查参数合法性
func (server *TCPServer) check_valid() {
	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Trace("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Trace("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
}

func (server *TCPServer) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	server.check_valid()

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
				return
			} else {
				log.Error("accept error: %v", err)
				continue
			}
		}

		//加入连接集合
		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Debug("too many connections")
			continue
		}
		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()

		server.wg.Add(1)

		tcpConn := newTCPConn(conn, server.PendingWriteNum, server.msgParser)
		agent := server.NewAgent(tcpConn)
		go func() {
			agent.Run()

			//从连接集合中删除
			tcpConn.Close()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()
			agent.OnClose()

			server.wg.Done()
		}()
	}
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
