package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"net"
)

type serverTCP struct {
	serverName string // 服务器名称
	ip         string // IP地址
	port       string // 服务端口
}

var serverTcp iface.IServer

func GetServerTCP() iface.IServer {
	if serverTcp == nil {
		serverTcp = &serverTCP{
			serverName: defaultServer.AppConf.AppName + "_tcp",
			ip:         defaultServer.AppConf.ServerTCP.Address,
			port:       defaultServer.AppConf.ServerTCP.Port,
		}
	}
	return serverTcp
}

func (s *serverTCP) GetServerName() string {
	return s.serverName
}

func (s *serverTCP) Start() {
	if s.port == "" {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	var (
		addr *net.TCPAddr
		tcp  *net.TCPListener
		conn *net.TCPConn
		err  error
	)

	// 1.获取TCP的Address
	addr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%s", s.ip, s.port))
	if err != nil {
		fmt.Printf("service startup failed %v\n", err)
		return
	}

	// 2.监听服务地址
	tcp, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		fmt.Printf("failed to listen to service address %v\n", err)
		return
	}
	defer func(tcp *net.TCPListener) {
		_ = tcp.Close()
	}(tcp)
	// 3.启动server网络连接业务
	for {
		// 等待客户端请求建立连接
		conn, err = tcp.AcceptTCP()
		if err != nil {
			fmt.Printf("accept tcp err %v\n", err)
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
		msgConn := NewConnectionTCP(s, conn)
		// 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	}
}
