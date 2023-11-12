package network

import (
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/logs"
	"github.com/gorilla/websocket"
	"net/http"
)

type ServerWS struct {
	Server
}

func NewServerWS() iface.IServer {
	s := &ServerWS{}
	s.serverName = config.GetGlobalObject().AppName + "_ws"
	s.ip = config.GetGlobalObject().ServerWS.Address
	s.port = config.GetGlobalObject().ServerWS.Port
	s.dataPacket = NewDataPack()
	return s
}

func (s *ServerWS) Start() {
	if s.isClose {
		s.isClose = false
		return
	}

	var upgrade = websocket.Upgrader{
		ReadBufferSize:  config.GetGlobalObject().MaxPackSize,
		WriteBufferSize: config.GetGlobalObject().MaxPackSize,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		// 服务关闭状态
		if s.isClose {
			_ = conn.Close()
			return
		}

		// 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= config.GetGlobalObject().MaxConn {
			_ = conn.Close()
			return
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionWS(s, conn)
		// 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
		// 建立连接成功
		logs.PrintLogInfo(fmt.Sprintf("成功建立新的客户端连接 -> %v connID - %v", msgConn.RemoteAddrStr(), msgConn.GetConnID()))
	})

	if certPath, keyPath := config.GetGlobalObject().ServerWS.TLSCertPath, config.GetGlobalObject().ServerWS.TLSKeyPath; certPath != "" && keyPath != "" {
		logs.PrintLogErr(http.ListenAndServeTLS(fmt.Sprintf("%s:%s", s.ip, s.port), certPath, keyPath, nil))
	} else {
		logs.PrintLogErr(http.ListenAndServe(fmt.Sprintf("%s:%s", s.ip, s.port), nil))
	}
}

func (s *ServerWS) Listen() bool {
	if config.GetGlobalObject().ServerWS.Port != "" {
		go s.Start()
		s.Server.Start()
		return true
	}
	return false
}
