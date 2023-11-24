package iface

import (
	"google.golang.org/protobuf/proto"
)

// 消息处理器
type IMsgHandler interface {
	// 获取已注册的协议列表
	GetApis() map[int32]IRouter
	// 异步处理消息
	DoMsgHandler(request IRequest)
	// 为消息注册解析体和处理函数
	AddRouter(msgId int32, msg INewMsgStructTemplate, handler IReceiveMsgHandler)
	// 将消息推入TaskQueue，等待Worker处理
	SendMsgToTaskQueue(request IRequest)
	// 设置过滤器
	SetFilter(fun IFilter)
	// 设置错误捕获
	SetErrCapture(fun IErrCapture)
}
type IFilter func(request IRequest, msgData proto.Message) bool
type IErrCapture func(request IRequest, r any)
