package rpc

import (
	"cell/common/mustang/rpc/wire"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
)

type clientCodec struct {
	conn io.ReadWriteCloser

	//TODO: clientCodec到底是每个call一个还是怎么回事？  respHeader不怕冲突吗？
	respHeader wire.ResponseHeader
	mutex      sync.Mutex
	pending    map[uint64]string
}

//NewClientCodec 初始化一个新的rpc.ClientCodec
func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{
		conn:    conn,
		pending: make(map[uint64]string),
	}
}

// WriteRequest must be safe for concurrent use by multiple goroutines.
//WriteRequest是不会并发的，但是有可能与ReadResponseHeader和ReadResponseBody并发
//ReadResponseHeader和ReadResponseBody有明确的先后顺序 且不会并发
func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	//TODO: API说这个函数必须协程安全，确定后面的writeRequest是协程安全的？
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	var request proto.Message
	if param != nil {
		var ok bool
		if request, ok = param.(proto.Message); !ok {
			return fmt.Errorf("rpc.ClientCodec.WriteRequest: %T does not implement proto.Message", param)
		}
	}

	err := writeRequest(c.conn, r.Seq, r.ServiceMethod, request)
	if err != nil {
		return err
	}

	return nil
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	header := wire.ResponseHeader{}
	err := readResponseHeader(c.conn, &header)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	r.Seq = header.Id
	r.Error = header.Error
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	c.respHeader = header
	return nil
}
func (c *clientCodec) ReadResponseBody(x interface{}) error {
	var response proto.Message
	if x != nil {
		var ok bool
		response, ok = x.(proto.Message)
		if !ok {
			return fmt.Errorf("rpc.ClientCodec.ReadResponseBody: %T does not implement proto.Message", x)
		}
	}

	err := readResponseBody(c.conn, &c.respHeader, response)
	if err != nil {
		return nil
	}

	c.respHeader = wire.ResponseHeader{}
	return nil
}

func (c *clientCodec) Close() error {
	return c.conn.Close()
}

//NewClient 初始化一个新的rpc.Client
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}

//Dial 连接
func Dial(address string) (*rpc.Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}

//DialTimeout 带有超时的尝试连接
func DialTimeout(address string, timeout time.Duration) (*rpc.Client, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}
