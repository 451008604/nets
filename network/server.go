package network

import (
	"fmt"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
	"net"

	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
)

// Server 定义Server服务类实现IServer接口
type Server struct {
	serverName  string                             // 服务器名称
	ipVersion   string                             // tcp4 or other
	ip          string                             // IP地址
	port        string                             // 服务端口
	msgHandler  iface.IMsgHandler                  // 当前Server的消息管理模块，用来绑定MsgId和对应的处理函数
	connMgr     iface.IConnManager                 // 当前Server的连接管理器
	onConnStart func(connection iface.IConnection) // 该Server连接创建时的Hook函数
	onConnStop  func(connection iface.IConnection) // 该Server连接断开时的Hook函数
	connID      uint32                             // 客户端连接自增ID
	dataPacket  iface.IDataPack                    // 数据拆包/封包工具
}

func NewServer() iface.IServer {
	s := &Server{
		serverName:  config.GetGlobalObject().Name,
		ipVersion:   "tcp4",
		ip:          config.GetGlobalObject().Host,
		port:        config.GetGlobalObject().TcpPort,
		msgHandler:  NewMsgHandler(),
		connMgr:     NewConnManager(),
		onConnStart: nil,
		onConnStop:  nil,
		connID:      0,
		dataPacket:  NewDataPack(),
	}
	return s
}

func (s *Server) Start() {
	// 开启一个go去做服务端Listener服务
	go func() {
		// 启动工作池等待接收请求数据
		s.msgHandler.StartWorkerPool()

		// 1.获取TCP的Address
		addr, err := net.ResolveTCPAddr(s.ipVersion, fmt.Sprintf("%s:%s", s.ip, s.port))
		if logs.PrintLogErr(err, "服务启动失败：") {
			return
		}

		// 2.监听服务地址
		tcp, err := net.ListenTCP(s.ipVersion, addr)
		if logs.PrintLogErr(err, "监听服务地址失败：") {
			return
		}

		// 3.启动server网络连接业务
		for {
			// 等待客户端建立请求连接
			var conn *net.TCPConn
			conn, err = tcp.AcceptTCP()
			if logs.PrintLogErr(err, "AcceptTCP ERR：") {
				continue
			}

			// 连接数量超过限制后，关闭新建立的连接
			if s.connMgr.Len() >= config.GetGlobalObject().MaxConn {
				_ = conn.Close()
				continue
			}

			// 自增connID
			s.connID = uint32(s.GetConnMgr().Len() + 1)
			// 建立连接成功
			logs.PrintLogInfo(fmt.Sprintf("成功建立新的客户端连接 -> %v connID - %v", conn.RemoteAddr().String(), s.connID))

			// 建立新的连接并监听客户端请求的消息
			dealConn := NewConnection(s, conn, s.connID, s.msgHandler)
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	logs.PrintLogInfo("服务关闭")

	s.connMgr.ClearConn()
}

func (s *Server) Listen() {
	s.Start()

	// 阻塞主线程
	select {}
}

func (s *Server) AddRouter(msgId pb.MessageID, msgStruct proto.Message, handler func(con iface.IConnection, message proto.Message)) {
	s.msgHandler.AddRouter(msgId, msgStruct, handler)
}

func (s *Server) GetConnMgr() iface.IConnManager {
	return s.connMgr
}

// Server连接创建时的Hook函数
func (s *Server) SetOnConnStart(f func(conn iface.IConnection)) {
	s.onConnStart = f
}

// 调用Server连接时的Hook函数
func (s *Server) CallbackOnConnStart(conn iface.IConnection) {
	if s.onConnStart != nil {
		s.onConnStart(conn)
	}
}

// Server连接断开时的Hook函数
func (s *Server) SetOnConnStop(f func(conn iface.IConnection)) {
	s.onConnStop = f
}

// 调用Server连接断开时的Hook函数
func (s *Server) CallbackOnConnStop(conn iface.IConnection) {
	if s.onConnStop != nil {
		s.onConnStop(conn)
	}
}

// 获取封包/拆包工具
func (s *Server) DataPacket() iface.IDataPack {
	return s.dataPacket
}
