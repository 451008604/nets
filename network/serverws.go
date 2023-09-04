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
	s.msgHandler = GetInstanceMsgHandler()
	s.connMgr = GetInstanceConnManager()
	s.dataPacket = NewDataPack()
	return s
}

func (s *ServerWS) Start() {
	var upgrade = websocket.Upgrader{
		ReadBufferSize:  1024 * 64,
		WriteBufferSize: 1024 * 64,
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
		msgConn := NewConnectionWS(s, conn, s.msgHandler)
		go msgConn.Start()
		// 建立连接成功
		logs.PrintLogInfo(fmt.Sprintf("成功建立新的客户端连接 -> %v connID - %v", msgConn.RemoteAddrStr(), msgConn.GetConnID()))
	})
	logs.PrintLogErr(http.ListenAndServe(fmt.Sprintf("%s:%s", s.ip, s.port), nil))
}

func (s *ServerWS) Listen() bool {
	if config.GetGlobalObject().HostWS != "" && config.GetGlobalObject().PortWS != "" {
		go s.Start()
		return true
	}
	return false
}
