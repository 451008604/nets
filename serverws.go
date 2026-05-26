package nets

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

type serverWS struct {
	serverName string // Server Name / 服务器名称
	ip         string // IP Address / IP地址
	port       int    // Service Port / 服务端口
}

var serverWs IServer

func GetServerWS() IServer {
	if serverWs == nil {
		serverWs = &serverWS{
			serverName: defaultServer.AppConf.AppName + "_ws",
			ip:         defaultServer.AppConf.ServerWS.Address,
			port:       defaultServer.AppConf.ServerWS.Port,
		}
	}
	return serverWs
}

func (s *serverWS) GetServerName() string {
	return s.serverName
}

func (s *serverWS) Start() {
	if s.port == 0 {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	var upgrade = websocket.Upgrader{
		ReadBufferSize:  defaultServer.AppConf.MaxPackSize,
		WriteBufferSize: 64 * 1024,
	}
	wsServer := http.NewServeMux()
	wsServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// Service Shutdown Status / 服务关闭状态
		if GetInstanceServerManager().IsClose() {
			_ = conn.Close()
			return
		}

		// Close new connections when count exceeds limit / 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= defaultServer.AppConf.MaxConn {
			_ = conn.Close()
			return
		}

		// Establish new connection and listen for client messages / 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionWS(s, conn)
		// Add new connection to unified connection manager / 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	})

	if certPath, keyPath := defaultServer.AppConf.ServerWS.TLSCertPath, defaultServer.AppConf.ServerWS.TLSKeyPath; certPath != "" && keyPath != "" {
		fmt.Printf("%v\n", http.ListenAndServeTLS(fmt.Sprintf("%s:%v", s.ip, s.port), certPath, keyPath, wsServer))
	} else {
		fmt.Printf("%v\n", http.ListenAndServe(fmt.Sprintf("%s:%v", s.ip, s.port), wsServer))
	}
}
