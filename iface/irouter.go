package iface

import (
	"google.golang.org/protobuf/proto"
)

type IRouter interface {
	// 设置消息体
	SetMsg(proto.Message)
	// 获取消息体
	GetMsg() proto.Message
	// 设置处理函数
	SetHandler(req IReceiveMsgHandler)
	// 执行处理函数
	RunHandler(request IRequest)
}

// IReceiveMsgHandler 逻辑处理模板函数
type IReceiveMsgHandler func(IConnection, proto.Message)
