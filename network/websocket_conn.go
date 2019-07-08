package network

import (
	"github.com/playnb/mustang/util"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type websocketConnSet map[*websocket.Conn]struct{}

//WSConn websocket链接对象
type WSConn struct {
	conn *websocket.Conn

	ImplConn
}

func newWSConn(conn *websocket.Conn, pendingWriteNum int) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.Init(wsConn, pendingWriteNum)
	wsConn.writeChan = make(chan []byte, pendingWriteNum)
	return wsConn
}

func (wsConn *WSConn) String() string {
	return fmt.Sprintf("[Conn:%s]", wsConn.ImplConn.String())
}

//CloseSession 关闭WebSocket
func (wsConn *WSConn) Close() error {
	return wsConn.conn.Close()
}

//WriteMsg 发送数据
func (wsConn *WSConn) WriteMsg(b *util.BuffData) error {
	wsConn.Write(b.GetPayload())
	return nil
}

func (wsConn *WSConn) write(b []byte, deadTime time.Time) (n int, err error) {
	if !deadTime.Equal(time.Unix(0, 0)) {
		wsConn.conn.SetWriteDeadline(deadTime)
	}
	err = wsConn.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

//ReadMsg 读取数据
func (wsConn *WSConn) ReadMsg() (*util.BuffData, error) {
	_, msg, err := wsConn.conn.ReadMessage()
	return util.MakeBuffDataBySlice(msg, 0), err
}

//LocalAddr 本地地址
func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

//RemoteAddr 远端地址
func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}
