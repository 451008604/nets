package network

import (
	"fmt"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/gorilla/websocket"
	"net/http"
)

type ServerWS struct {
	serverName  string                             // 服务器名称
	ip          string                             // IP地址
	port        string                             // 服务端口
	msgHandler  iface.IMsgHandler                  // 当前Server的消息管理模块，用来绑定MsgId和对应的处理函数
	connMgr     iface.IConnManager                 // 当前Server的连接管理器
	dataPacket  iface.IDataPack                    // 数据拆包/封包工具
	onConnStart func(connection iface.IConnection) // 该Server连接创建时的Hook函数
	onConnStop  func(connection iface.IConnection) // 该Server连接断开时的Hook函数
}

func NewServerWS() iface.IServer {
	s := &ServerWS{
		serverName: config.GetGlobalObject().Name + "_ws",
		ip:         config.GetGlobalObject().HostWS,
		port:       config.GetGlobalObject().PortWS,
		msgHandler: NewMsgHandler(),
		connMgr:    NewConnManager(),
		dataPacket: NewDataPack(),
	}
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
		if s.connMgr.Len() >= config.GetGlobalObject().MaxConn {
			_ = conn.Close()
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnection(s, conn, s.msgHandler)
		go msgConn.Start()

		// // Go程序开启一个新的Goroutines，确保我们可以同时处理多个WebSocket连接
		// for {
		// 	// 读取从WebSocket接收的消息
		// 	messageType, p, err := conn.ReadMessage()
		// 	if err != nil {
		// 		fmt.Println(err)
		// 		return
		// 	}
		//
		// 	// 将消息打印到控制台
		// 	fmt.Println(string(p))
		//
		// 	// 将接收到的消息回传给WebSocket
		// 	if err := conn.WriteMessage(messageType, p); err != nil {
		// 		fmt.Println(err)
		// 		return
		// 	}
		// }
	})
	logs.PrintLogErr(http.ListenAndServe(fmt.Sprintf("%s:%s", s.ip, s.port), nil))
}

func (s *ServerWS) Stop() {
	// TODO implement me
	panic("implement me")
}

func (s *ServerWS) Listen() bool {
	if config.GetGlobalObject().HostWS != "" && config.GetGlobalObject().PortWS != "" {
		go s.Start()
		return true
	}
	return false
}

func (s *ServerWS) AddRouter(msgId pb.MessageID, msgStruct iface.INewMsgStructTemplate, handler iface.IReceiveMsgHandler) {
	// TODO implement me
	panic("implement me")
}

func (s *ServerWS) GetConnMgr() iface.IConnManager {
	// TODO implement me
	panic("implement me")
}

func (s *ServerWS) SetOnConnStart(f func(conn iface.IConnection)) {
	// TODO implement me
	panic("implement me")
}

func (s *ServerWS) CallbackOnConnStart(conn iface.IConnection) {
	// TODO implement me
	panic("implement me")
}

func (s *ServerWS) SetOnConnStop(f func(conn iface.IConnection)) {
	// TODO implement me
	panic("implement me")
}

func (s *ServerWS) CallbackOnConnStop(conn iface.IConnection) {
	// TODO implement me
	panic("implement me")
}

func (s *ServerWS) DataPacket() iface.IDataPack {
	// TODO implement me
	panic("implement me")
}
