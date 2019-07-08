package rpc

import (
	"github.com/playnb/mustang/rpc/wire"
	"errors"
	"fmt"
	"io"
	"net/rpc"
	"sync"

	"github.com/gogo/protobuf/proto"
)

type serverCodec struct {
	conn io.ReadWriteCloser

	reqHeader wire.RequestHeader

	mutex   sync.Mutex // protects seq, pending
	seq     uint64
	pending map[uint64]uint64
}

/*
通过观察 rpc.Client.input() 函数
接收数据ReadRequestHeader,ReadRequestBody是成对调用的
*/
//NewServerCodec new rpc.ServerCodec
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{
		conn:    conn,
		pending: make(map[uint64]uint64),
	}
}

func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	header := wire.RequestHeader{}
	err := readRequestHeader(s.conn, &header)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	s.seq++
	s.pending[s.seq] = header.Id
	r.ServiceMethod = header.Method
	r.Seq = s.seq //这里存的server的seq
	s.mutex.Unlock()

	s.reqHeader = header

	return nil
}

func (s *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		//WriteRequest失败了
		return nil
	}

	request, ok := x.(proto.Message)
	if !ok {
		return fmt.Errorf("rpc.ServerCodec.ReadRequestBody: %T does not implement proto.Message", x)
	}

	err := readRequestBody(s.conn, &s.reqHeader, request)
	if err != nil {
		return nil
	}

	s.reqHeader = wire.RequestHeader{}
	return nil
}

//WriteResponse 不会并发
func (s *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		if response, ok = x.(proto.Message); !ok {
			if _, ok = x.(struct{}); !ok {
				s.mutex.Lock()
				delete(s.pending, r.Seq)
				s.mutex.Unlock()
				return fmt.Errorf("rpc.ServerCodec.WriteResponse: %T does not implement proto.Message", x)
			}
		}
	}

	s.mutex.Lock()
	id, ok := s.pending[r.Seq] //获取客户端的seq
	if !ok {
		s.mutex.Unlock()
		return errors.New("rpc: invalid sequence number in response")
	}
	delete(s.pending, r.Seq)
	s.mutex.Unlock()

	err := writeResponse(s.conn, id, r.Error, response)
	if err != nil {
		return err
	}

	return nil
}

func (s *serverCodec) Close() error {
	return s.conn.Close()
}

//ServeConn rpc.ServeConn
func ServeConn(conn io.ReadWriteCloser) {
	rpc.ServeCodec(NewServerCodec(conn))
}
