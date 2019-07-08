package network

import (
	"github.com/playnb/mustang/util"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	//"time"
	//	"net"
	//	"sync"

	//	"common/mustang/log"
)

// -----------------------------------
// |flag(1)| len(3) |      data      |
// -----------------------------------

type MsgParser struct {
	LittleEndian bool
	LenMsgLen    uint32
	MaxMsgLen    uint32
	MinMsgLen    uint32
	MsgMask      uint32

	ReadState   uint32
	CurReadLen  uint32
	CurReadFlag uint32
}

func (parser *MsgParser) String() string {
	return fmt.Sprintf("LittleEndian:%v MsgMask:%d MinMsgLen:%d MaxMsgLen:%d LenMsgLen:%d ReadState:%d CurReadLen:%d CurReadFlag:%d",
		parser.LittleEndian,
		parser.MsgMask,
		parser.MinMsgLen,
		parser.MaxMsgLen,
		parser.LenMsgLen,
		parser.ReadState,
		parser.CurReadLen,
		parser.CurReadFlag)
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.LittleEndian = false
	p.LenMsgLen = 4
	p.MaxMsgLen = math.MaxUint32 - 4
	p.MinMsgLen = 1
	p.MsgMask = 0xFFFFFF
	p.ReadState = 0
	return p
}

//获取数据长度
func (p *MsgParser) readMsgLength(conn io.Reader, msgLen *uint32, msgFlag *uint32) error {
	*msgLen = 0
	*msgFlag = 0

	//TODO 从缓冲区读取数据
	bufMsgLen := make([]byte, p.LenMsgLen)

	p.ReadState = 11
	//io.ReadFull 一直读取,直到缓冲区读满为止
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return err
	}
	var msgHead = uint32(0)
	if p.LittleEndian {
		msgHead = binary.LittleEndian.Uint32(bufMsgLen)
	} else {
		msgHead = binary.BigEndian.Uint32(bufMsgLen)
	}

	*msgLen = msgHead % p.MsgMask
	*msgFlag = msgHead / p.MsgMask

	p.ReadState = 12
	p.CurReadLen = *msgLen
	p.CurReadFlag = *msgFlag

	if *msgLen > p.MaxMsgLen {
		return errors.New("message too long")
	}
	if *msgLen < p.MinMsgLen {
		return errors.New("message too short " + strconv.Itoa((int)(*msgLen)))
	}
	return nil
}

//读取数据
func (p *MsgParser) Read(conn io.Reader) (*util.BuffData, error) {
	p.ReadState = 10
	var msgLen uint32
	var msgFlag uint32
	if err := p.readMsgLength(conn, &msgLen, &msgFlag); err != nil {
		return nil, err
	}

	p.ReadState = 20

	/*
		//TODO 从缓冲区读取数据
		msgData := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, msgData); err != nil {
			p.ReadState = 30
			return nil, err
		}
	*/
	//msgData:= util.MakeBuffData(int(msgLen))
	msgData := util.DefaultPool().Get(int(msgLen))
	if _, err := io.ReadFull(conn, msgData.GetPayload()); err != nil {
		p.ReadState = 30
		return nil, err
	}

	p.ReadState = 40
	return msgData, nil
}

func (p *MsgParser) Write(conn *TCPConn, data *util.BuffData) (*util.BuffData, error) {
	//__ft3 := util.NewFunctionTime("on_MsgParser", 1)
	// get len
	var msgLen uint32
	var msgFlag uint32
	var msgHead uint32

	msgFlag = 0
	msgLen = uint32(data.Size())

	// check len
	if msgLen > p.MaxMsgLen {
		return nil, errors.New("message too long")
	} else if msgLen < p.MinMsgLen {
		return nil, errors.New("message too short")
	}
	if !data.ChangeIndex(-int(p.LenMsgLen)) {
		return nil, errors.New("BuffData.Index too short")
	}

	msgHead = msgFlag*p.MsgMask + msgLen

	// 写消息大小
	if p.LittleEndian {
		binary.LittleEndian.PutUint32(data.GetPayload(), msgHead)
	} else {
		binary.BigEndian.PutUint32(data.GetPayload(), msgHead)
	}

	// 写消息数据
	//__ft3.End()

	//conn.Write(msg)
	return data, nil
}
