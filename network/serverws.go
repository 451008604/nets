package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
	"net/http"
)

type serverWS struct {
	serverName string // 服务器名称
	ip         string // IP地址
	port       string // 服务端口
}

var serverWs iface.IServer

func GetServerWS() iface.IServer {
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
	if s.port == "" {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	var upgrade = websocket.Upgrader{
		ReadBufferSize:  defaultServer.AppConf.MaxPackSize,
		WriteBufferSize: defaultServer.AppConf.MaxPackSize,
	}
	wsServer := http.NewServeMux()
	wsServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// 服务关闭状态
		if GetInstanceServerManager().IsClose() {
			_ = conn.Close()
			return
		}

		// 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= defaultServer.AppConf.MaxConn {
			_ = conn.Close()
			return
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionWS(s, conn)
		// 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	})

	if certPath, keyPath := defaultServer.AppConf.ServerWS.TLSCertPath, defaultServer.AppConf.ServerWS.TLSKeyPath; certPath != "" && keyPath != "" {
		fmt.Printf("%v\n", http.ListenAndServeTLS(fmt.Sprintf("%s:%s", s.ip, s.port), certPath, keyPath, wsServer))
	} else {
		fmt.Printf("%v\n", http.ListenAndServe(fmt.Sprintf("%s:%s", s.ip, s.port), wsServer))
	}
}
