package network

import (
	"github.com/playnb/mustang/log"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//wsHandler 实现net.http.handler接口,每次HTTP请求的时候被调用,用于产生链接
type wsHandler struct {
	server     *WSServer
	upgrader   websocket.Upgrader
	conns      websocketConnSet
	mutexConns sync.Mutex
	wg         sync.WaitGroup
}

func (handler *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer log.PrintPanicStack()
	log.Trace("New from %s(url:%s)", r.RemoteAddr, r.RequestURI)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("upgrade error: %v", err)
		return
	}

	handler.mutexConns.Lock()
	if handler.conns == nil {
		handler.mutexConns.Unlock()
		conn.Close()
		log.Error("upgrade error: nil <%v>", conn)
		return
	}
	if len(handler.conns) >= handler.server.MaxConnNum {
		handler.mutexConns.Unlock()
		conn.Close()
		log.Error("too many connections <%v>", conn)
		return
	}
	handler.conns[conn] = struct{}{}
	handler.mutexConns.Unlock()

	wsConn := newWSConn(conn, handler.server.PendingWriteNum)
	wsConn.SetAgentObj(handler.server.NewAgent(wsConn), handler.server)

	handler.wg.Add(1)
	defer handler.wg.Done()

	log.Trace("Begin Loop %s(url:%s)[%s]", r.RemoteAddr, r.RequestURI, wsConn)

	go wsConn.SendLoop()
	wsConn.RecvLoop()

	// cleanup
	handler.server.OnTCPConnClose(wsConn)
	log.Trace("Close %s(url:%s)[%s]", r.RemoteAddr, r.RequestURI, wsConn)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//WSServer websocket服务器
type WSServer struct {
	opts            *ServerOptions
	Addr            string
	MaxConnNum      int //最大连接数量
	PendingWriteNum int //最大消息队列数量
	NewAgent        func(*WSConn) IAgent
	HTTPTimeout     time.Duration //http超时时间(ms)
	ln              net.Listener
	handler         *wsHandler
	wg              sync.WaitGroup
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
		server: server,
		conns:  make(websocketConnSet),
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

//CloseSession 关闭服务器
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

func (server *WSServer) OnTCPConnClose(tcpConn IConn) {
	server.handler.mutexConns.Lock()
	delete(server.handler.conns, tcpConn.(*WSConn).conn)
	server.handler.mutexConns.Unlock()
}

func NewWebsocketServer(opts *ServerOptions) Server {
	if opts == nil {
		panic(errors.New("创建服务器的opts不能为空"))
	}
	if opts.NewAgent == nil {
		panic(errors.New("创建服务器的NewAgent不能为空"))
	}
	server := &WSServer{}
	server.opts = opts
	server.Addr = opts.BindAddr
	log.Trace("%s 初始化Websocket服务 %s", server.opts.Name, server.opts.BindAddr)

	server.MaxConnNum = opts.MaxConnNum
	server.PendingWriteNum = opts.PendingWriteNum
	server.NewAgent = func(conn *WSConn) IAgent { return opts.NewAgent(conn) }
	return server
}
