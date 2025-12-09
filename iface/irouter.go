package iface

import (
	"google.golang.org/protobuf/proto"
)

// 定义协议路由模板
type IRouter interface {
	// 设置消息体
	SetMsg(msgTemplate INewMsgStructTemplate)
	// 获取新的空消息结构体
	GetNewMsg() proto.Message
	// 设置处理函数
	SetHandler(msgHandler IReceiveMsgHandler)
	// 执行处理函数
	RunHandler(conn IConnection, message proto.Message)
}
type IReceiveMsgHandler func(conn IConnection, message proto.Message)
type INewMsgStructTemplate func() proto.Message
