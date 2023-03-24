package iface

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

// 定义服务器接口
type IServer interface {
	// 启动服务器
	Start()
	// 停止服务器
	Stop()
	// 开启业务服务
	Listen()
	// 给当前服务注册一个路由业务方法，提供给客户端连接使用
	AddRouter(msgId pb.MessageID, msgStruct proto.Message, handler func(con IConnection, message proto.Message))
	// 获取连接管理器
	GetConnMgr() IConnManager
	// Server连接创建时的Hook函数
	SetOnConnStart(func(conn IConnection))
	// 调用Server连接时的Hook函数
	CallbackOnConnStart(conn IConnection)
	// Server连接断开时的Hook函数
	SetOnConnStop(func(conn IConnection))
	// 调用Server连接断开时的Hook函数
	CallbackOnConnStop(conn IConnection)
	// 获取封包/拆包工具
	DataPacket() IDataPack
}
