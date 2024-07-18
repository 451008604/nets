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
	// 将任务放入执行队列，等待Worker处理
	PushInTaskQueue(task ITaskTemplate)
	// 设置过滤器
	SetFilter(fun IFilter)
	// 设置错误捕获
	SetErrCapture(fun IErrCapture)
}
type IFilter func(conn IConnection, msgData proto.Message) bool
type IErrCapture func(conn IConnection, r any)

// 任务模版
type ITaskTemplate interface {
	TaskHandler(conn IConnection)
}
