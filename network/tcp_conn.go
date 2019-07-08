package network

import (
	//	"encoding/binary"
	//	"errors"
	//	"io"
	//	"math"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/playnb/mustang/log"
	"github.com/playnb/mustang/util"
	"fmt"
)

type ConnSet map[net.Conn]struct{}

var sequenceTCPConnID = uint64(1)

type ITcpConnOwner interface {
	OnTCPConnClose(tcpConn *TCPConn)
}

//TCPConn TCP连接
type TCPConn struct {
	connMutex  sync.Mutex
	conn       net.Conn
	writeChan  chan *util.BuffData
	closeFlag  bool //关闭标识
	msgParser  *MsgParser
	sequenceID uint64

	agent IAgent
	owner ITcpConnOwner

	exitChan       chan bool
	canDropMessage bool
}

//创建TCP连接
func newTCPConn(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.closeFlag = false
	tcpConn.writeChan = make(chan *util.BuffData, pendingWriteNum)
	tcpConn.exitChan = make(chan bool, 1)
	tcpConn.msgParser = msgParser
	tcpConn.sequenceID = atomic.AddUint64(&sequenceTCPConnID, 1)

	return tcpConn
}

func (tcpConn *TCPConn) String() string {
	return fmt.Sprintf("[TCPConn:%d]", tcpConn.sequenceID)
}

//SetAgentObj 设置Agent
func (tcpConn *TCPConn) SetAgentObj(agent IAgent, owner ITcpConnOwner) {
	tcpConn.owner = owner
	tcpConn.agent = agent
	tcpConn.agent.SetConn(tcpConn)
	tcpConn.agent.ConnectFunc()
}

func (tcpConn *TCPConn) SetCloseFlag() {
}
func (tcpConn *TCPConn) GetExitChan() chan bool {
	return nil
}
func (tcpConn *TCPConn) GetCloseFlag() bool {
	return false
}
func (tcpConn *TCPConn) Close() error {
	return nil
}
func (tcpConn *TCPConn) write(b []byte, deadTime time.Time) (n int, err error) {
	return n, nil
}

//SendLoop 每个链接都有一个goroutine处理发送消息
func (tcpConn *TCPConn) SendLoop() {
	log.Debug("[%s] 开始发送数据goroutine", tcpConn)
	needBreak := false
	for !needBreak {
		//__ft2 := util.NewFunctionTime("SendLoop", 100)
		select {
		case b, ok := <-tcpConn.writeChan:
			if ok == false {
				log.Debug("%s 关闭发送channel writeChan", tcpConn)
				needBreak = true
				break
			}
			if b == nil {
				log.Debug("%s nil消息 主动结束发送数据goroutine", tcpConn)
				//needBreak = true
				break
			}

			//TODO 写之前判断连接状态,不能写也不直接被break
			if tcpConn.closeFlag == true {
				log.Error("%s TCPConn Error:"+"向关闭的端口写数据", tcpConn)
				//needBreak = true
				break
			}

			tcpConn.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			len, err := tcpConn.conn.Write(b.GetPayload())
			b.Release()
			if len > 0 {
			}
			//if tcpConn.agent != nil && tcpConn.agent.GetDebug() {
			//log.Ltp("[%s] TCPConn: 数据确实发送 %d", tcpConn, len)
			//}

			if err != nil {
				log.Error(tcpConn.String() + "==================> Error:" + err.Error())
				//time.Sleep(time.Second * 60)
				//log.Trace(tcpConn.String() + "==================> Error:" + err.Error())
				//needBreak = true
				//tcpConn.Terminate()
				tcpConn.conn.Close()
				break
			}
		}
		//__ft2.End()

		if needBreak == true {
			break
		}
	}
	log.Debug("[%s] 发送数据goroutine结束, 关闭连接", tcpConn)
	tcpConn.Terminate()
}

//RecvLoop 每个链接都有一个接收数据的gorountinue
func (tcpConn *TCPConn) RecvLoop() {
	defer log.PrintPanicStack()

	log.Debug("[%s] 开始接收数据goroutine", tcpConn)
	tcpConn.agent.Run(tcpConn.agent)
	log.Debug("[%s] 接收goroutine结束, 关闭连接", tcpConn)

	tcpConn.Terminate()

	tcpConn.onClose()
}

//Terminate 关闭连接(请求)
func (tcpConn *TCPConn) Terminate() {
	log.Debug("%s 调用 Terminate", tcpConn)

	tcpConn.connMutex.Lock()
	defer tcpConn.connMutex.Unlock()
	if tcpConn.closeFlag {
		log.Debug("%s tcpConn.closeFlag 为真，不走TCPConn Terminate", tcpConn)
		return
	}
	log.Debug("%s tcpConn Terminate, closeFlag(%v)", tcpConn, tcpConn.closeFlag)
	tcpConn.closeFlag = true
	tcpConn.conn.Close()
	close(tcpConn.exitChan)
}

//onClose 关闭连接(动作)
func (tcpConn *TCPConn) onClose() {
	func() {
		tcpConn.connMutex.Lock()
		defer tcpConn.connMutex.Unlock()

		if tcpConn.agent != nil {
			log.Debug("tcpConn onClose %v", tcpConn.agent)
		} else {
			log.Debug("tcpConn onClose")
		}
		tcpConn.closeFlag = true
		close(tcpConn.writeChan)
	}()

	if tcpConn.agent.CloseFunc != nil {
		tcpConn.agent.CloseFunc()
	}
	tcpConn.agent = nil
	tcpConn.conn.Close()
	tcpConn.owner.OnTCPConnClose(tcpConn)
}

//Write 写入数据
func (tcpConn *TCPConn) Write(b *util.BuffData) {
	tcpConn.connMutex.Lock()
	defer tcpConn.connMutex.Unlock()
	if tcpConn.closeFlag == true {
		log.Debug("[%s] TCPConn:Write 向关闭的Chan写数据", tcpConn)
		return
	}
	//if tcpConn.agent != nil && tcpConn.agent.GetDebug() {
	//	log.Debug("[%s] TCPConn:Write 发送数据1 %d (len:%d,cap:%d)", tcpConn, len(b), len(tcpConn.writeChan), cap(tcpConn.writeChan))
	//}

	msgCount := len(tcpConn.writeChan)
	if msgCount > cap(tcpConn.writeChan)/2+10 {
		log.Debug("[%s] TCPConn:Write 有消息累积(%d)", tcpConn, msgCount)
		if tcpConn.canDropMessage {
			log.Debug("[%s] TCPConn:Write 抛弃消息(%d)", tcpConn, msgCount)
			return
		}
	}
	__ft3 := util.NewFunctionTime("on_Write", 1)
	tcpConn.writeChan <- b
	__ft3.End()
	//if tcpConn.agent != nil && tcpConn.agent.GetDebug() {
	//	log.Debug("[%s] TCPConn:Write 发送数据2", tcpConn)
	//}
}

//Read 从conn中读取数据
func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

//LocalAddr 获取本地地址
func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

//RemoteAddr 获取远端地址
func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

//ReadMsg 读取Msg
func (tcpConn *TCPConn) ReadMsg() (*util.BuffData, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

//WriteMsg 写入Msg
func (tcpConn *TCPConn) WriteMsg(data *util.BuffData) error {
	if tcpConn.closeFlag {
		return errors.New("向一个已经关闭的TCPConn写数据 " + tcpConn.String())
	}
	buf, err := tcpConn.msgParser.Write(tcpConn, data)
	if err != nil {
		return err
	}
	tcpConn.Write(buf)
	return nil
}

func (tcpConn *TCPConn) WriteMsgDirectly(data *util.BuffData) error {
	if tcpConn.closeFlag {
		return errors.New("向一个已经关闭的TCPConn写数据 " + tcpConn.String())
	}
	buf, err := tcpConn.msgParser.Write(tcpConn, data)
	if err != nil {
		return err
	}

	tcpConn.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	len, err := tcpConn.conn.Write(buf.GetPayload())
	if len > 0 {
	}
	if err != nil {
		log.Error(tcpConn.String() + "==================> Error:" + err.Error())
		tcpConn.conn.Close()
	}
	return err
}
