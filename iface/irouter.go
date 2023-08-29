package iface

import (
	"google.golang.org/protobuf/proto"
)

type IRouter interface {
	// 设置消息体
	SetMsg(msgStructTemplate INewMsgStructTemplate)
	// 获取新的空消息结构体
	GetNewMsg() proto.Message
	// 设置处理函数
	SetHandler(req IReceiveMsgHandler)
	// 执行处理函数
	RunHandler(request IRequest, message proto.Message)
}

// IReceiveMsgHandler 逻辑处理模板函数
type IReceiveMsgHandler func(con IConnection, message proto.Message)

// INewMsgStructTemplate 空消息结构体模版
type INewMsgStructTemplate func() proto.Message
