package network

import (
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	"sync"
)

type Server struct {
	serverName string             // 服务器名称
	ip         string             // IP地址
	port       string             // 服务端口
	msgHandler iface.IMsgHandler  // 当前Server的消息管理模块，用来绑定MsgId和对应的处理函数
	connMgr    iface.IConnManager // 当前Server的连接管理器
	dataPacket iface.IDataPack    // 数据拆包/封包工具
}

var ServerWaitFlag = &sync.WaitGroup{}

func (s *Server) GetServerName() string {
	return s.serverName
}

func (s *Server) Start() {
}

func (s *Server) Stop() {
	logs.PrintLogInfo("服务关闭")

	s.GetConnMgr().ClearConn()
	ServerWaitFlag.Done()
}

func (s *Server) Listen() bool {
	return false
}

func (s *Server) GetConnMgr() iface.IConnManager {
	return s.connMgr
}

// 获取封包/拆包工具
func (s *Server) DataPacket() iface.IDataPack {
	return s.dataPacket
}
