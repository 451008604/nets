package network

import (
	"fmt"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"net"
)

// ServerTCP 定义Server服务类实现IServer接口
type ServerTCP struct {
	serverName string             // 服务器名称
	ip         string             // IP地址
	port       string             // 服务端口
	msgHandler iface.IMsgHandler  // 当前Server的消息管理模块，用来绑定MsgId和对应的处理函数
	connMgr    iface.IConnManager // 当前Server的连接管理器
	dataPacket iface.IDataPack    // 数据拆包/封包工具
}

func NewServerTCP() iface.IServer {
	s := &ServerTCP{
		serverName: config.GetGlobalObject().Name + "_tcp",
		ip:         config.GetGlobalObject().HostTCP,
		port:       config.GetGlobalObject().PortTCP,
		msgHandler: NewMsgHandler(),
		connMgr:    NewConnManager(),
		dataPacket: NewDataPack(),
	}
	return s
}

func (s *ServerTCP) Start() {
	// 启动工作池等待接收请求数据
	s.msgHandler.StartWorkerPool()

	var (
		addr *net.TCPAddr
		tcp  *net.TCPListener
		conn *net.TCPConn
		err  error
	)

	// 1.获取TCP的Address
	addr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%s", s.ip, s.port))
	if logs.PrintLogErr(err, "服务启动失败：") {
		return
	}

	// 2.监听服务地址
	tcp, err = net.ListenTCP("tcp4", addr)
	if logs.PrintLogErr(err, "监听服务地址失败：") {
		return
	}

	// 3.启动server网络连接业务
	for {
		// 等待客户端请求建立连接
		conn, err = tcp.AcceptTCP()
		if logs.PrintLogErr(err, "AcceptTCP ERR：") {
			continue
		}

		// 连接数量超过限制后，关闭新建立的连接
		if s.GetConnMgr().Len() >= config.GetGlobalObject().MaxConn {
			_ = conn.Close()
			continue
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnection(s, conn, s.msgHandler)
		go msgConn.Start()
	}
}

func (s *ServerTCP) Stop() {
	logs.PrintLogInfo("服务关闭")

	s.GetConnMgr().ClearConn()
}

func (s *ServerTCP) Listen() bool {
	if config.GetGlobalObject().HostTCP != "" && config.GetGlobalObject().PortTCP != "" {
		go s.Start()
		return true
	}
	return false
}

func (s *ServerTCP) AddRouter(msgId pb.MessageID, msgStruct iface.INewMsgStructTemplate, handler iface.IReceiveMsgHandler) {
	s.msgHandler.AddRouter(msgId, msgStruct, handler)
}

func (s *ServerTCP) GetConnMgr() iface.IConnManager {
	return s.connMgr
}

// 获取封包/拆包工具
func (s *ServerTCP) DataPacket() iface.IDataPack {
	return s.dataPacket
}
