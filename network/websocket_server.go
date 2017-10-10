package network

import (
	"github.com/playnb/mustang/log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//wsHandler 实现net.http.handler接口,每次HTTP请求的时候被调用,用于产生链接
type wsHandler struct {
	maxConnNum      int
	pendingWriteNum int
	newAgent        func(*WSConn) IAgent
	upgrader        websocket.Upgrader
	conns           websocketConnSet
	mutexConns      sync.Mutex
	wg              sync.WaitGroup
}

func (handler *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer log.PrintPanicStack()
	log.Trace("new connect from %s(url:%s)", r.RemoteAddr, r.RequestURI)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("upgrade error: %v", err)
		return
	}

	handler.wg.Add(1)
	defer handler.wg.Done()

	handler.mutexConns.Lock()
	if handler.conns == nil {
		handler.mutexConns.Unlock()
		conn.Close()
		log.Error("upgrade error: nil <%v>", conn)
		return
	}
	if len(handler.conns) >= handler.maxConnNum {
		handler.mutexConns.Unlock()
		conn.Close()
		log.Error("too many connections <%v>", conn)
		return
	}
	handler.conns[conn] = struct{}{}
	handler.mutexConns.Unlock()

	wsConn := newWSConn(conn, handler.pendingWriteNum)
	agent := handler.newAgent(wsConn)
	agent.Run(agent)

	// cleanup
	agent.CloseFunc()
	wsConn.Close()
	handler.mutexConns.Lock()
	delete(handler.conns, conn)
	handler.mutexConns.Unlock()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//WSServer websocket服务器
type WSServer struct {
	Addr            string
	MaxConnNum      int //最大连接数量
	PendingWriteNum int //最大消息队列数量
	NewAgent        func(*WSConn) IAgent
	HTTPTimeout     time.Duration //http超时时间(ms)
	ln              net.Listener
	handler         *wsHandler
}

//Start 启动一个HTTP服务器,在Addr指定的地址
func (server *WSServer) Start() {
	//监听端口
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	server.checkValid()

	server.ln = ln
	server.handler = &wsHandler{
		maxConnNum:      server.MaxConnNum,
		pendingWriteNum: server.PendingWriteNum,
		newAgent:        server.NewAgent,
		conns:           make(websocketConnSet),
		upgrader: websocket.Upgrader{ //Upgrader对象
			ReadBufferSize:   64 * 1024,                                  //读缓冲
			WriteBufferSize:  64 * 1024,                                  //写缓冲
			HandshakeTimeout: server.HTTPTimeout,                         //设置各种超时
			CheckOrigin:      func(_ *http.Request) bool { return true }, //忽略跨域访问限制
		},
	}

	//创建http服务器
	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        server.handler,
		ReadTimeout:    server.HTTPTimeout, //设置各种超时
		WriteTimeout:   server.HTTPTimeout, //设置各种超时
		MaxHeaderBytes: 1024,
	}

	//启动HTTP服务器
	go httpServer.Serve(ln)
}

//Close 关闭服务器
func (server *WSServer) Close() {
	//关闭所有websocket连接
	server.handler.mutexConns.Lock()
	for conn := range server.handler.conns {
		conn.Close()
	}
	server.handler.conns = make(websocketConnSet)
	server.handler.mutexConns.Unlock()

	//关闭TCP连接
	server.ln.Close()

	//等待所有websocket退出
	server.handler.wg.Wait()
}

//checkValid 配置合理性检查
func (server *WSServer) checkValid() {
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
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Trace("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}
}
