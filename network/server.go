package network

type Server interface {
	Start()
}

type ServerOptions struct {
	Name            string
	BindAddr        string
	MaxConnNum      int //最大连接数量
	PendingWriteNum int //最大消息队列数量
	NewAgent        func(IConn) IAgent
}
