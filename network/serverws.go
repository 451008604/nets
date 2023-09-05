package network

import (
	"fmt"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/gorilla/websocket"
	"net/http"
)

type ServerWS struct {
	Server
}

func NewServerWS() iface.IServer {
	s := &ServerWS{}
	s.serverName = config.GetGlobalObject().Name + "_ws"
	s.ip = config.GetGlobalObject().HostWS
	s.port = config.GetGlobalObject().PortWS
	s.connMgr = GetInstanceConnManager()
	s.dataPacket = NewDataPack()
	return s
}

func (s *ServerWS) Start() {
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

		// 连接数量超过限制后，关闭新建立的连接
		if s.GetConnMgr().Len() >= config.GetGlobalObject().MaxConn {
			_ = conn.Close()
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionWS(s, conn)
		go msgConn.Start(msgConn.StartWriter)
		// 建立连接成功
		logs.PrintLogInfo(fmt.Sprintf("成功建立新的客户端连接 -> %v connID - %v", msgConn.RemoteAddrStr(), msgConn.GetConnID()))
	})

	if certPath, keyPath := config.GetGlobalObject().TLSCertPath, config.GetGlobalObject().TLSKeyPath; certPath != "" && keyPath != "" {
		logs.PrintLogErr(http.ListenAndServeTLS(fmt.Sprintf("%s:%s", s.ip, s.port), certPath, keyPath, nil))
	} else {
		logs.PrintLogErr(http.ListenAndServe(fmt.Sprintf("%s:%s", s.ip, s.port), nil))
	}
}

func (s *ServerWS) Listen() bool {
	if config.GetGlobalObject().PortWS != "" {
		go s.Start()
		return true
	}
	return false
}
