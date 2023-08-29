package iface

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"net"
)

type IConnection interface {
	// 启动连接
	Start()
	// 停止连接
	Stop()

	// 从当前连接获取原始的Socket TCPConn
	GetTCPConnection() *net.TCPConn
	// 获取当前连接ID
	GetConnID() int
	// 获取客户端地址信息
	RemoteAddr() net.Addr

	// 发送消息给客户端（无缓冲）
	SendMsg(msgId pb.MessageID, data []byte)
	// 发送消息给客户端（有缓冲）
	SendBuffMsg(msgId pb.MessageID, data []byte)

	// 设置连接属性
	SetProperty(key string, value interface{})
	// 获取连接属性
	GetProperty(key string) (value interface{})
	// 删除连接属性
	RemoveProperty(key string)
}
