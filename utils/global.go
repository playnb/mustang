package utils

import (
	"github.com/playnb/mustang/network/protobuf"
)

//定义一些全局变量在这里

//关闭的命令
var CloseSig = make(chan bool)

//基于protobuf的处理器
var ProtobufProcess = protobuf.NewProcessor()
