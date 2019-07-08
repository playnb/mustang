package main

import (
	protorpc "cell/common/mustang/rpc"
	"cell/common/mustang/rpc/example/testrpc"
	"io"
	"log"
	"net"
	"net/rpc"
)

type EchoService interface {
	Echo(in *testrpc.RequestEcho, out *testrpc.ReplyEcho) error
}

//NOT NECESSARY
func AcceptEchoServiceClient(lis net.Listener, x EchoService) {
	srv := rpc.NewServer()
	if err := srv.RegisterName("EchoService", x); err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("lis.Accept(): %v\n", err)
		}
		go srv.ServeCodec(protorpc.NewServerCodec(conn))
	}
}

//NOT NECESSARY
func RegisterEchoService(srv *rpc.Server, x EchoService) error {
	if err := srv.RegisterName("EchoService", x); err != nil {
		return err
	}
	return nil
}

//NOT NECESSARY
func NewEchoServiceServer(x EchoService) *rpc.Server {
	srv := rpc.NewServer()
	if err := srv.RegisterName("EchoService", x); err != nil {
		log.Fatal(err)
	}
	return srv
}

func ListenAndServeEchoService(addr string, x EchoService) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer lis.Close()

	srv := rpc.NewServer()
	if err := srv.RegisterName("EchoService", x); err != nil {
		return err
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("lis.Accept(): %v\n", err)
		}
		go srv.ServeCodec(protorpc.NewServerCodec(conn))
	}
}

type EchoServiceClient struct {
	*rpc.Client
	//Echo(in *testrpc.RequestEcho, out *testrpc.ReplyEcho)
}

//NOT NECESSARY
func NewEchoServiceClient(conn io.ReadWriteCloser) *EchoServiceClient {
	c := rpc.NewClientWithCodec(protorpc.NewClientCodec(conn))
	return &EchoServiceClient{c}
}

func (c *EchoServiceClient) Echo(in *testrpc.RequestEcho) (out *testrpc.ReplyEcho, err error) {
	if in == nil {
		in = new(testrpc.RequestEcho)
	}
	out = new(testrpc.ReplyEcho)
	if err = c.Call("EchoService.Echo", in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func DialEchoService(addr string) (*EchoServiceClient, error) {
	c, err := protorpc.Dial(addr)
	if err != nil {
		return nil, err
	}
	return &EchoServiceClient{c}, nil
}
