package iface

import (
	"google.golang.org/protobuf/proto"
)

// 消息处理器
type IMsgHandler interface {
	// 获取已注册的协议列表
	GetApis() map[int32]IRouter
	// 为消息注册解析体和处理函数
	AddRouter(msgId int32, msg INewMsgStructTemplate, handler IReceiveMsgHandler)
	// 设置过滤器
	SetFilter(fun IFilter)
	// 获取过滤器
	GetFilter() IFilter
	// 设置错误捕获
	SetErrCapture(fun IErrCapture)
	// 获取错误捕获
	GetErrCapture(conn IConnection)
}
type IFilter func(conn IConnection, msgData proto.Message) bool
type IErrCapture func(conn IConnection, r any)
