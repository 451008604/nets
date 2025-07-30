package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"github.com/xtaci/kcp-go"
	"net"
)

type serverKCP struct {
	serverName string
	ip         string
	port       string
}

var serverKcp iface.IServer

func GetServerKCP() iface.IServer {
	if serverKcp == nil {
		serverKcp = &serverKCP{
			serverName: defaultServer.AppConf.AppName + "_kcp",
			ip:         defaultServer.AppConf.ServerKCP.Address,
			port:       defaultServer.AppConf.ServerKCP.Port,
		}
	}
	return serverKcp
}

func (s *serverKCP) GetServerName() string {
	return s.serverName
}

func (s *serverKCP) Start() {
	if s.port == "" {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	var conn net.Conn

	listener, err := kcp.Listen(":" + s.port)
	if err != nil {
		fmt.Printf("service startup failed %v\n", err)
		return
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	for {
		conn, err = listener.Accept()
		if err != nil {
			fmt.Printf("accept kcp err %v\n", err)
			continue
		}

		// 服务关闭状态
		if GetInstanceServerManager().IsClose() {
			_ = conn.Close()
			continue
		}

		// 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= defaultServer.AppConf.MaxConn {
			_ = conn.Close()
			continue
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionKCP(s, conn)
		// 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	}
}
