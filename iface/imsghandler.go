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

	RegisterTaskHandler(taskType WorkerTaskType, handler func(conn IConnection))
	// 将消息推入TaskQueue，等待Worker处理
	PushInTaskQueue(taskType WorkerTaskType, conn IConnection)
	// 设置过滤器
	SetFilter(fun IFilter)
	// 设置错误捕获
	SetErrCapture(fun IErrCapture)
}
type IFilter func(conn IConnection, msgData proto.Message) bool
type IErrCapture func(conn IConnection, r any)

// 工作任务类型
type WorkerTaskType string

const (
	SysReaderMessage = "SysReaderMessage" // 接受到数据
)
