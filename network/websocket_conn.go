package network

import (
	"github.com/playnb/mustang/log"
	"errors"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

type websocketConnSet map[*websocket.Conn]struct{}

//WSConn websocket链接对象
type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	writeChan chan []byte
	closeFlag bool
}

func newWSConn(conn *websocket.Conn, pendingWriteNum int) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.writeChan = make(chan []byte, pendingWriteNum)

	go func() {
		for b := range wsConn.writeChan {
			if b == nil {
				break
			}

			err := wsConn.conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				log.Error(err.Error())
				break
			}
		}

		conn.Close()
		wsConn.Lock()
		wsConn.closeFlag = true
		wsConn.Unlock()
	}()

	return wsConn
}

func (wsConn *WSConn) onDestroy() {
	wsConn.conn.Close()
	close(wsConn.writeChan)
	wsConn.closeFlag = true
}

//Close 关闭WebSocket
func (wsConn *WSConn) Close() {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}
	wsConn.onDestroy()
}

//WriteMsg 发送数据
func (wsConn *WSConn) WriteMsg(b []byte) error {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		log.Trace("%s 向关闭的WebSocket写数据")
		return errors.New("向关闭的WebSocket写数据")
	}
	/*
		if len(wsConn.writeChan) == cap(wsConn.writeChan) {
			log.Debug("close conn: channel full")
			return
		}
	*/
	wsConn.writeChan <- b
	return nil
}

//ReadMsg 读取数据
func (wsConn *WSConn) ReadMsg() ([]byte, error) {
	_, msg, err := wsConn.conn.ReadMessage()
	return msg, err
}

//LocalAddr 本地地址
func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

//RemoteAddr 远端地址
func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}
