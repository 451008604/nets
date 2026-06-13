package nets

import (
	"fmt"
	"net"
	"sync"
)

type serverTCP struct {
	serverName string // Server Name / 服务器名称
	ip         string // IP Address / IP地址
	port       int    // Service Port / 服务端口
}

var (
	serverTcp     IServer
	serverTcpOnce sync.Once
)

func GetServerTCP() IServer {
	serverTcpOnce.Do(func() {
		serverTcp = &serverTCP{
			serverName: defaultServer.AppConf.AppName + "_tcp",
			ip:         defaultServer.AppConf.ServerTCP.Address,
			port:       defaultServer.AppConf.ServerTCP.Port,
		}
	})
	return serverTcp
}

func (s *serverTCP) GetServerName() string {
	return s.serverName
}

func (s *serverTCP) Start() {
	if s.port == 0 {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	var (
		addr *net.TCPAddr
		tcp  *net.TCPListener
		conn *net.TCPConn
		err  error
	)

	addr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%v", s.ip, s.port))
	if err != nil {
		fmt.Printf("service startup failed %v\n", err)
		return
	}

	tcp, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		fmt.Printf("failed to listen to service address %v\n", err)
		return
	}
	defer func(tcp *net.TCPListener) {
		_ = tcp.Close()
	}(tcp)

	// Close the listener on shutdown so the blocking Accept returns immediately
	// 退出时关闭监听器，使阻塞的 Accept 立即返回
	go func() {
		<-serverCtx.Done()
		_ = tcp.Close()
	}()

	for {
		// Wait for client request to establish connection / 等待客户端请求建立连接
		conn, err = tcp.AcceptTCP()
		if err != nil {
			select {
			case <-serverCtx.Done():
				return
			default:
			}
			fmt.Printf("accept tcp err %v\n", err)
			continue
		}

		// Service Shutdown Status / 服务关闭状态
		if GetInstanceServerManager().IsClose() {
			_ = conn.Close()
			continue
		}

		// Close new connections when count exceeds limit / 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= defaultServer.AppConf.MaxConn {
			_ = conn.Close()
			continue
		}

		// Establish new connection and listen for client messages / 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionTCP(s, conn)
		// Add new connection to unified connection manager / 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	}
}
