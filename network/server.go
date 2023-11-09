package network

import (
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/logs"
	"sync"
)

type Server struct {
	serverName string             // 服务器名称
	ip         string             // IP地址
	port       string             // 服务端口
	isClose    bool               // 服务是否已关闭
	connMgr    iface.IConnManager // 当前Server的连接管理器
	dataPacket iface.IDataPack    // 数据拆包/封包工具
}

var ServerWaitFlag = &sync.WaitGroup{}

func (s *Server) GetServerName() string {
	return s.serverName
}

func (s *Server) Start() {
	ServerWaitFlag.Add(1)
}

func (s *Server) Stop() {
	logs.PrintLogInfo("服务关闭")

	s.GetConnMgr().ClearConn()
	s.isClose = true
	ServerWaitFlag.Done()
}

func (s *Server) Listen() bool {
	s.isClose = false
	return false
}

func (s *Server) GetConnMgr() iface.IConnManager {
	return s.connMgr
}

// 获取封包/拆包工具
func (s *Server) DataPacket() iface.IDataPack {
	return s.dataPacket
}
