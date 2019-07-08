package network

import (
	"github.com/playnb/mustang/util"
	"net"
	"time"
)

type IConn interface {
	String() string
	GetCloseFlag() bool
	SetCloseFlag()
	RemoteAddr() net.Addr
	GetExitChan() chan bool

	Terminate()

	SendLoop()
	RecvLoop()

	ReadMsg() (*util.BuffData, error)
	WriteMsg(data *util.BuffData) error
	//Read(b []byte) (n int, err error)
	write(b []byte, deadTime time.Time) (n int, err error)
	Close() error
}
