package network

//WSAgent websocket的agent
type WSAgent struct {
	ImplAgent
}

func (a *WSAgent) SetConn(conn IConn) {
	a.conn = conn.(*WSConn)
}

func (a *WSAgent) String() string {
	return a.conn.String()
}
