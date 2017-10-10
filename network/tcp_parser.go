package network

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strconv"
	//	"net"
	//	"sync"

	//	"github.com/playnb/mustang/log"
)

// -----------------------------------
// |flag(1)| len(3) |      data      |
// -----------------------------------

type MsgParser struct {
	littleEndian bool
	lenMsgLen    uint32
	maxMsgLen    uint32
	minMsgLen    uint32
	msgMask      uint32
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.littleEndian = false
	p.lenMsgLen = 4
	p.maxMsgLen = math.MaxUint32 - 4
	p.minMsgLen = 1
	p.msgMask = 0xFFFFFF
	return p
}

//获取数据长度
func (p *MsgParser) readMsgLength(conn *TCPConn, msgLen *uint32, msgFlag *uint32) error {
	*msgLen = 0
	*msgFlag = 0

	//TODO 从缓冲区读取数据
	bufMsgLen := make([]byte, p.lenMsgLen)

	//io.ReadFull 一直读取,直到缓冲区读满为止
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return err
	}
	var msgHead  = uint32(0)
	if p.littleEndian {
		msgHead = binary.LittleEndian.Uint32(bufMsgLen)
	} else {
		msgHead = binary.BigEndian.Uint32(bufMsgLen)
	}

	*msgLen = msgHead % p.msgMask
	*msgFlag = msgHead / p.msgMask

	if *msgLen > p.maxMsgLen {
		return errors.New("message too long")
	}
	if *msgLen < p.minMsgLen {
		return errors.New("message too short " + strconv.Itoa((int)(*msgLen)))
	}
	return nil
}

//读取数据
func (p *MsgParser) Read(conn *TCPConn) ([]byte, error) {
	var msgLen uint32
	var msgFlag uint32
	if err := p.readMsgLength(conn, &msgLen, &msgFlag); err != nil {
		return nil, err
	}

	//TODO 从缓冲区读取数据
	msgData := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err
	}

	return msgData, nil
}

func (p *MsgParser) Write(conn *TCPConn, data []byte) error {
	// get len
	var msgLen uint32
	var msgFlag uint32
	var msgHead uint32

	msgFlag = 0
	msgLen = uint32(len(data))

	// check len
	if msgLen > p.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return errors.New("message too short")
	}

	msgHead = msgFlag*p.msgMask + msgLen

	//TODO 从缓冲区读取数据
	msg := make([]byte, uint32(p.lenMsgLen)+msgLen)

	// 写消息大小
	if p.littleEndian {
		binary.LittleEndian.PutUint32(msg, msgHead)
	} else {
		binary.BigEndian.PutUint32(msg, msgHead)
	}

	// 写消息数据
	copy(msg[p.lenMsgLen:], data)

	conn.Write(msg)

	return nil
}
