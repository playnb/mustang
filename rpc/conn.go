package rpc

import (
	"github.com/playnb/mustang/rpc/wire"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
)

//发送一帧数据
func sendFrame(w io.Writer, data []byte) (err error) {
	var size [binary.MaxVarintLen64]byte //数据大小

	if data == nil || len(data) == 0 {
		n := binary.PutUvarint(size[:], uint64(0))
		if err = write(w, size[:n], false); err != nil { //发送数据大小
			return
		}
		return
	}

	n := binary.PutUvarint(size[:], uint64(len(data)))
	if err = write(w, size[:n], false); err != nil { //发送数据大小
		return
	}
	if err = write(w, data, false); err != nil { //发送数据
		return
	}

	return
}

//读取一帧数据
func recvFrame(r io.Reader) (data []byte, err error) {
	size, err := readUvarint(r) //读取需要接收的数据大小
	if err != nil {
		return nil, err
	}
	if size != 0 {
		data = make([]byte, size)
		if err = read(r, data); err != nil { //接受数据
			return nil, err
		}
	}
	return data, nil
}

func write(w io.Writer, data []byte, onePacket bool) error {
	//TODO: onePacket? 一次发送或者分多次尝试发送
	if onePacket {
		if _, err := w.Write(data); err != nil {
			return err
		}
		return nil
	}
	for index := 0; index < len(data); {
		n, err := w.Write(data[index:])
		if err != nil {
			if nerr, ok := err.(net.Error); !ok || !nerr.Temporary() {
				return err
			}
		}
		index += n
	}
	return nil
}

//按照binary.Uvarint改编
/*
func Uvarint(buf []byte) (uint64, int) {
	var x uint64
	var s uint
	for i, b := range buf {
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0, -(i + 1) // overflow
			}
			return x | uint64(b)<<s, i + 1
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, 0
}
*/
func readUvarint(r io.Reader) (uint64, error) {
	var x uint64
	var s uint
	for i := 0; ; i++ {
		var b byte
		b, err := readByte(r)
		if err != nil {
			return 0, err
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0, errors.New("rpc: varint overflows a 64-bit integer")
			}
			return x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}

//io.ReadFull
func read(r io.Reader, data []byte) error {
	for index := 0; index < len(data); {
		n, err := r.Read(data[index:])
		if err != nil {
			if nerr, ok := err.(net.Error); !ok || !nerr.Temporary() {
				return err
			}
		}
		index += n
	}
	return nil
}

func readByte(r io.Reader) (c byte, err error) {
	data := make([]byte, 1)
	if err = read(r, data); err != nil {
		return 0, err
	}
	c = data[0]
	return
}

//发送请求
func writeRequest(w io.Writer, id uint64, method string, request proto.Message) error {
	//准备request数据
	var pbRequest []byte
	if request != nil {
		var err error
		pbRequest, err = proto.Marshal(request)
		if err != nil {
			return err
		}
	}

	//用snappy压缩数据
	compressedPbRequest := snappy.Encode(nil, pbRequest)

	//准备request头部数据
	header := &wire.RequestHeader{
		Id:                         id,
		Method:                     method,
		RawRequestLen:              uint32(len(pbRequest)),
		SnappyCompressedRequestLen: uint32(len(compressedPbRequest)),
		Checksum:                   crc32.ChecksumIEEE(compressedPbRequest),
	}
	pbHeader, err := proto.Marshal(header)
	if err != err {
		return err
	}

	if len(pbHeader) > int(wire.Const_MaxRequestHeaderLen) {
		return fmt.Errorf("rpc.writeRequest: header larger than MaxRequestHeaderLen: %d", len(pbHeader))
	}

	//发送头部(一帧)
	if err := sendFrame(w, pbHeader); err != nil {
		return err
	}

	//发送数据(一帧)
	if err := sendFrame(w, compressedPbRequest); err != nil {
		return err
	}

	return nil
}

//读取request头部数据
func readRequestHeader(r io.Reader, header *wire.RequestHeader) error {
	//读取一帧数据
	pbHeader, err := recvFrame(r)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(pbHeader, header)
	if err != nil {
		return err
	}

	return nil
}

//读取request数据
func readRequestBody(r io.Reader, header *wire.RequestHeader, request proto.Message) error {
	//读取一帧数据
	compressedPbRequest, err := recvFrame(r)
	if err != nil {
		return err
	}

	//检查校验和
	if crc32.ChecksumIEEE(compressedPbRequest) != header.Checksum {
		return fmt.Errorf("rpc.readRequestBody: unexpected checksum")
	}

	//解压缩数据
	pbRequest, err := snappy.Decode(nil, compressedPbRequest)
	if err != nil {
		return err
	}

	//校验原始数据大小
	if uint32(len(pbRequest)) != header.RawRequestLen {
		return fmt.Errorf("rpc.readRequestBody: Unexcpeted header.RawRequestLen")
	}

	//发序列化request
	if request != nil {
		err = proto.Unmarshal(pbRequest, request)
		if err != nil {
			return err
		}
	}

	return nil
}

//=========================== response处理同request
func writeResponse(w io.Writer, id uint64, serr string, response proto.Message) error {
	if serr != "" {
		response = nil
	}

	var pbResponse []byte
	if response != nil {
		var err error
		pbResponse, err = proto.Marshal(response)
		if err != nil {
			return err
		}
	}
	compressedPbResponse := snappy.Encode(nil, pbResponse)

	header := &wire.ResponseHeader{
		Id:                          id,
		Error:                       serr,
		RawResponseLen:              uint32(len(pbResponse)),
		SnappyCompressedResponseLen: uint32(len(compressedPbResponse)),
		Checksum:                    crc32.ChecksumIEEE(compressedPbResponse),
	}
	pbHeader, err := proto.Marshal(header)
	if err != nil {
		return err
	}

	if err := sendFrame(w, pbHeader); err != nil {
		return err
	}

	if err := sendFrame(w, compressedPbResponse); err != nil {
		return err
	}

	return nil
}

func readResponseHeader(r io.Reader, header *wire.ResponseHeader) error {
	pbHeader, err := recvFrame(r)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(pbHeader, header)
	if err != nil {
		return err
	}

	return nil
}

func readResponseBody(r io.Reader, header *wire.ResponseHeader, response proto.Message) error {
	compressedPbResponse, err := recvFrame(r)
	if err != nil {
		return err
	}

	if crc32.ChecksumIEEE(compressedPbResponse) != header.Checksum {
		return fmt.Errorf("rpc.readResponseBody: unexpected checksum")
	}

	pbResponse, err := snappy.Decode(nil, compressedPbResponse)
	if err != nil {
		return err
	}

	if uint32(len(pbResponse)) != header.RawResponseLen {
		return fmt.Errorf("rpc.readResponseBody: Unexcpeted header.RawResponseLen")
	}

	if response != nil {
		err = proto.Unmarshal(pbResponse, response)
		if err != nil {
			return err
		}
	}

	return nil
}
