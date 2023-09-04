package iface

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
)

type IConnection interface {
	// 启动连接
	Start()
	// 停止连接
	Stop()

	// 启动接收消息协程
	StartReader()
	// 启动发送消息协程
	StartWriter()

	// 获取当前连接ID
	GetConnID() int
	// 获取客户端地址信息
	RemoteAddrStr() string

	// 发送消息给客户端
	SendMsg(msgId pb.MessageID, data []byte)

	// 设置连接属性
	SetProperty(key string, value interface{})
	// 获取连接属性
	GetProperty(key string) (value interface{})
	// 删除连接属性
	RemoveProperty(key string)
}
