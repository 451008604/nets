package iface

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
)

// 定义服务器接口
type IServer interface {
	// 启动服务器
	Start()
	// 停止服务器
	Stop()
	// 开启业务服务
	Listen() bool
	// 给当前服务注册一个路由业务方法，提供给客户端连接使用
	AddRouter(msgId pb.MessageID, msgStruct INewMsgStructTemplate, handler IReceiveMsgHandler)
	// 获取连接管理器
	GetConnMgr() IConnManager
	// 获取封包/拆包工具
	DataPacket() IDataPack
}
