package network

import (
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
	"net/http"
)

type serverWS struct {
	server
}

func NewServerWS() iface.IServer {
	s := &serverWS{}
	s.serverName = config.GetServerConf().AppName + "_ws"
	s.ip = config.GetServerConf().ServerWS.Address
	s.port = config.GetServerConf().ServerWS.Port
	s.dataPacket = NewDataPack()
	return s
}

func (s *serverWS) Start() {
	if s.isClose {
		s.isClose = false
		return
	}

	var upgrade = websocket.Upgrader{
		ReadBufferSize:  config.GetServerConf().MaxPackSize,
		WriteBufferSize: config.GetServerConf().MaxPackSize,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			_ = conn.Close()
			return
		}

		// 服务关闭状态
		if s.isClose {
			_ = conn.Close()
			return
		}

		// 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= config.GetServerConf().MaxConn {
			_ = conn.Close()
			return
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionWS(s, conn)
		// 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	})

	if certPath, keyPath := config.GetServerConf().ServerWS.TLSCertPath, config.GetServerConf().ServerWS.TLSKeyPath; certPath != "" && keyPath != "" {
		fmt.Printf("%v\n", http.ListenAndServeTLS(fmt.Sprintf("%s:%s", s.ip, s.port), certPath, keyPath, nil))
	} else {
		fmt.Printf("%v\n", http.ListenAndServe(fmt.Sprintf("%s:%s", s.ip, s.port), nil))
	}
}

func (s *serverWS) Listen() bool {
	if config.GetServerConf().ServerWS.Port != "" {
		go s.Start()
		s.server.Start()
		fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)
		return true
	}
	return false
}
