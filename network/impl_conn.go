package network

import (
	"github.com/playnb/mustang/log"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type IConnOwner interface {
	OnTCPConnClose(tcpConn IConn)
}

type ImplConn struct {
	connMutex sync.Mutex
	//conn       net.Conn
	writeChan  chan []byte
	closeFlag  bool //关闭标识
	sequenceID uint64

	agent    IAgent
	owner    IConnOwner
	realConn IConn

	exitChan chan bool
}

func (tcpConn *ImplConn) String() string {
	return fmt.Sprintf("[%s	SID:%d]", tcpConn.realConn.RemoteAddr().String(), tcpConn.sequenceID)
}

var connSequenceID uint64

func (tcpConn *ImplConn) Init(realConn IConn, pendingWriteNum int) {
	tcpConn.closeFlag = false
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.exitChan = make(chan bool, 1)
	tcpConn.sequenceID = atomic.AddUint64(&sequenceTCPConnID, 1)
	tcpConn.realConn = realConn
}

func (tcpConn *ImplConn) GetExitChan() chan bool {
	return tcpConn.exitChan
}

//SetAgentObj 设置Agent
func (tcpConn *ImplConn) SetAgentObj(agent IAgent, owner IConnOwner) {
	tcpConn.owner = owner
	tcpConn.agent = agent
	tcpConn.agent.SetConn(tcpConn.realConn)
	tcpConn.agent.ConnectFunc()
}

//SendLoop 每个链接都有一个goroutine处理发送消息
func (tcpConn *ImplConn) SendLoop() {
	log.Debug("[%s] 开始发送数据goroutine", tcpConn)
	needBreak := false
	for !needBreak {
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

			len, err := tcpConn.realConn.write(b, time.Now().Add(time.Second*30))
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
				tcpConn.realConn.Close()
				break
			}
		}

		if needBreak == true {
			break
		}
	}
	log.Debug("[%s] 发送数据goroutine结束, 关闭连接", tcpConn)
	tcpConn.Terminate()
}

//RecvLoop 每个链接都有一个接收数据的gorountinue
func (tcpConn *ImplConn) RecvLoop() {
	defer log.PrintPanicStack()

	log.Debug("[%s] 开始接收数据goroutine", tcpConn.String())
	tcpConn.agent.Run(tcpConn.agent)
	log.Debug("[%s] 接收goroutine结束, 关闭连接", tcpConn.String())

	tcpConn.Terminate()

	tcpConn.onClose()
}

//Terminate 关闭连接(请求)
func (tcpConn *ImplConn) Terminate() {
	log.Debug("%s 调用 Terminate", tcpConn)

	tcpConn.connMutex.Lock()
	defer tcpConn.connMutex.Unlock()
	if tcpConn.closeFlag {
		log.Debug("%s tcpConn.closeFlag 为真，不走TCPConn Terminate", tcpConn)
		return
	}
	log.Debug("%s tcpConn Terminate, closeFlag(%v)", tcpConn, tcpConn.closeFlag)
	tcpConn.closeFlag = true
	tcpConn.realConn.Close()
	close(tcpConn.exitChan)
}

//onClose 关闭连接(动作)
func (tcpConn *ImplConn) onClose() {
	func() {
		tcpConn.connMutex.Lock()
		defer tcpConn.connMutex.Unlock()

		if tcpConn.agent != nil {
			log.Debug("tcpConn onClose %s", tcpConn.agent)
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
	tcpConn.realConn.Close()
	tcpConn.owner.OnTCPConnClose(tcpConn.realConn)
}

//Write 写入数据
func (tcpConn *ImplConn) Write(b []byte) {
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
	if msgCount > 250 {
		log.Debug("[%s] TCPConn:Write 有消息累积(%d)", tcpConn, msgCount)
	}
	tcpConn.writeChan <- b
	//if tcpConn.agent != nil && tcpConn.agent.GetDebug() {
	//	log.Debug("[%s] TCPConn:Write 发送数据2", tcpConn)
	//}
}

func (tcpConn *ImplConn) GetCloseFlag() bool {
	return tcpConn.closeFlag
}

func (tcpConn *ImplConn) SetCloseFlag() {
	tcpConn.closeFlag = true
}

/*
//Read 从conn中读取数据
func (tcpConn *ImplConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

//LocalAddr 获取本地地址
func (tcpConn *ImplConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

//RemoteAddr 获取远端地址
func (tcpConn *ImplConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

//ReadMsg 读取Msg
func (tcpConn *ImplConn) ReadMsg() ([]byte, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

//WriteMsg 写入Msg
func (tcpConn *ImplConn) WriteMsg(data []byte) error {
	if tcpConn.closeFlag {
		return errors.New("向一个已经关闭的TCPConn写数据 " + tcpConn.String())
	}
	return tcpConn.msgParser.Write(tcpConn, data)
}
*/
