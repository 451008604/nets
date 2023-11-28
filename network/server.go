package network

import (
	"github.com/451008604/nets/iface"
	"sync"
)

type server struct {
	serverName string             // 服务器名称
	ip         string             // IP地址
	port       string             // 服务端口
	isClose    bool               // 服务是否已关闭
	connMgr    iface.IConnManager // 当前Server的连接管理器
	dataPacket iface.IDataPack    // 数据拆包/封包工具
}

var ServerWaitFlag = &sync.WaitGroup{}

func (s *server) GetServerName() string {
	return s.serverName
}

func (s *server) Start() {
	ServerWaitFlag.Add(1)
}

func (s *server) Stop() {
	GetInstanceConnManager().ClearConn()
	s.isClose = true
	ServerWaitFlag.Done()
}

func (s *server) Listen() bool {
	s.isClose = false
	return false
}

func (s *server) DataPacket() iface.IDataPack {
	return s.dataPacket
}
