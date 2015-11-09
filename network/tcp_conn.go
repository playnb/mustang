package network

import (
	"TestGo/mustang/log"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"sync"
)

type ConnSet map[net.Conn]struct{}

// -----------------
// | len(4) | data |
// -----------------

//解析TCP数据包
type MsgParser struct {
	littleEndian bool
	lenMsgLen    uint32
	maxMsgLen    uint32
	minMsgLen    uint32
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.littleEndian = false
	p.lenMsgLen = 4
	p.maxMsgLen = math.MaxUint32 - 4
	p.minMsgLen = 1
	return p
}

//获取数据长度
func (p *MsgParser) readMsgLeg(conn *TCPConn, msgLen *uint32) error {
	*msgLen = 0
	bufMsgLen := make([]byte, p.lenMsgLen)
	//io.ReadFull 一直读取,知道缓冲区读满为止
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return err
	}
	if p.littleEndian {
		*msgLen = binary.LittleEndian.Uint32(bufMsgLen)
	} else {
		*msgLen = binary.BigEndian.Uint32(bufMsgLen)
	}
	if *msgLen > p.maxMsgLen {
		return errors.New("message too long")
	}
	if *msgLen < p.minMsgLen {
		return errors.New("message too short")
	}
	return nil
}

//读取数据
func (p *MsgParser) Read(conn *TCPConn) ([]byte, error) {
	var msgLen uint32
	if err := p.readMsgLeg(conn, &msgLen); err != nil {
		return nil, err
	}

	msgData := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err
	}
	return msgData, nil
}

func (p *MsgParser) Write(conn *TCPConn, args ...[]byte) error {
	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// check len
	if msgLen > p.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return errors.New("message too short")
	}

	msg := make([]byte, uint32(p.lenMsgLen)+msgLen)

	// 写消息大小
	if p.littleEndian {
		binary.LittleEndian.PutUint32(msg, msgLen)
	} else {
		binary.BigEndian.PutUint32(msg, msgLen)
	}

	// 写消息数据
	index := p.lenMsgLen
	// 拼接多个数据块
	for i := 0; i < len(args); i++ {
		copy(msg[index:], args[i])
		index += uint32(len(args[i]))
	}

	conn.Write(msg)

	return nil
}

//////////////////////////////////////////////////////////////
type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool //关闭标识
	msgParser *MsgParser
}

func newTCPConn(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser

	//每个链接都有一个goroutine处理发送消息
	go func() {
		for b := range tcpConn.writeChan {
			if b == nil {
				break
			}

			_, err := conn.Write(b)
			if err != nil {
				break
			}
		}

		conn.Close()
		tcpConn.Lock()
		tcpConn.closeFlag = true
		tcpConn.Unlock()
	}()

	return tcpConn
}

func (tcpConn *TCPConn) doDestroy() {
	tcpConn.conn.(*net.TCPConn).SetLinger(0)
	tcpConn.conn.Close()
	close(tcpConn.writeChan)
	tcpConn.closeFlag = true
}

func (tcpConn *TCPConn) Destroy() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}
	tcpConn.doDestroy()
}

func (tcpConn *TCPConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}
	tcpConn.doWrite(nil)
	tcpConn.closeFlag = true
}

func (tcpConn *TCPConn) doWrite(b []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doDestroy()
		return
	}
	tcpConn.writeChan <- b
}

func (tcpConn *TCPConn) Write(b []byte) {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag || b == nil {
		return
	}
	tcpConn.doWrite(b)
}

func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

func (tcpConn *TCPConn) ReadMsg() ([]byte, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

func (tcpConn *TCPConn) WriteMsg(args ...[]byte) error {
	return tcpConn.msgParser.Write(tcpConn, args...)
}
