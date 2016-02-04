package utils

import ()

//定义一些全局变量在这里
//服务类型
const (
	GateService = iota
	LogicService
	SuperService
)

//获取唯一ID的类别
const (
	SnowflakeSystemWork = 0 //唯一ID的work类型
)

const (
	SnowflakeCatalogAuth = 1 //获取认证流程的唯一ID
)

func GetServiceName(service_type int) string {
	switch service_type {
	case GateService:
		return "GateService"
	case LogicService:
		return "LogicService"
	case SuperService:
		return "SuperService"
	}
	return "UnknowService"
}

//关闭的命令
var CloseSig = make(chan bool)

//========================================
//Super服务的地址
const SuperRpcAddr = "127.0.0.1:19090"
