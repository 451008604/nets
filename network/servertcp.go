package network

import (
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/logs"
	"net"
)

type ServerTCP struct {
	Server
}

func NewServerTCP() iface.IServer {
	s := &ServerTCP{}
	s.serverName = config.GetGlobalObject().AppName + "_tcp"
	s.ip = config.GetGlobalObject().ServerTCP.Address
	s.port = config.GetGlobalObject().ServerTCP.Port
	s.connMgr = GetInstanceConnManager()
	s.dataPacket = NewDataPack()
	return s
}

func (s *ServerTCP) Start() {
	if s.isClose {
		s.isClose = false
		return
	}

	var (
		addr *net.TCPAddr
		tcp  *net.TCPListener
		conn *net.TCPConn
		err  error
	)

	// 1.获取TCP的Address
	addr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%s", s.ip, s.port))
	if logs.PrintLogErr(err, "服务启动失败：") {
		return
	}

	// 2.监听服务地址
	tcp, err = net.ListenTCP("tcp4", addr)
	if logs.PrintLogErr(err, "监听服务地址失败：") {
		return
	}

	// 3.启动server网络连接业务
	for {
		// 等待客户端请求建立连接
		conn, err = tcp.AcceptTCP()
		if logs.PrintLogErr(err, "AcceptTCP ERR：") {
			continue
		}

		// 服务关闭状态
		if s.isClose {
			_ = conn.Close()
			continue
		}

		// 连接数量超过限制后，关闭新建立的连接
		if s.GetConnMgr().Len() >= config.GetGlobalObject().MaxConn {
			_ = conn.Close()
			continue
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionTCP(s, conn)
		go msgConn.Start(msgConn.StartReader, msgConn.StartWriter)
		// 建立连接成功
		logs.PrintLogInfo(fmt.Sprintf("成功建立新的客户端连接 -> %v connID - %v", msgConn.RemoteAddrStr(), msgConn.GetConnID()))
	}
}

func (s *ServerTCP) Listen() bool {
	if config.GetGlobalObject().ServerTCP.Port != "" {
		go s.Start()
		s.Server.Start()
		return true
	}
	return false
}
