package nets

import (
	"fmt"
	"github.com/xtaci/kcp-go"
	"net"
)

type serverKCP struct {
	serverName string
	ip         string
	port       int
}

var serverKcp IServer

func GetServerKCP() IServer {
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
	if s.port == 0 {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	listener, err := kcp.Listen(fmt.Sprintf("%s:%v", s.ip, s.port))
	if err != nil {
		fmt.Printf("service startup failed %v\n", err)
		return
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	// Close the listener on shutdown so the blocking Accept returns immediately
	// 退出时关闭监听器，使阻塞的 Accept 立即返回
	go func() {
		<-serverCtx.Done()
		_ = listener.Close()
	}()

	var conn net.Conn
	for {
		conn, err = listener.Accept()
		if err != nil {
			select {
			case <-serverCtx.Done():
				return
			default:
			}
			fmt.Printf("accept kcp err %v\n", err)
			continue
		}

		if sess, ok := conn.(*kcp.UDPSession); ok {
			sess.SetNoDelay(1, 10, 2, 1)
			sess.SetWindowSize(256, 256)
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
		msgConn := NewConnectionKCP(s, conn)
		// Add new connection to unified connection manager / 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	}
}
