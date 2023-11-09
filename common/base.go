package common

import (
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/network"
)

var module = &staticModule{}

type staticModule struct {
	serverTCP iface.IServer // 服务进程模块
	serverWS  iface.IServer // 服务进程模块
}

func init() {
	module.serverTCP = network.NewServerTCP()
	module.serverWS = network.NewServerWS()
}

func GetServerTCP() iface.IServer {
	return module.serverTCP
}

func GetServerWS() iface.IServer {
	return module.serverWS
}
