package iface

import (
	"google.golang.org/protobuf/proto"
)

type IConnection interface {
	// 启动连接(通过connmanager调用)
	Start(readerHandler func() bool, writerHandler func(data []byte) bool)
	// 停止连接(通过connmanager调用)
	Stop()

	// 启动接收消息协程
	StartReader() bool
	// 启动发送消息协程
	StartWriter(data []byte) bool
	// 执行任务
	DoTask(task func())

	// 获取当前连接Id
	GetConnId() int
	// 获取客户端地址信息
	RemoteAddrStr() string
	// 获取连接是否已关闭
	IsClose() bool

	// 发送消息给客户端
	SendMsg(msgId int32, msgData proto.Message)

	// 设置连接属性
	SetProperty(key IConnProperty, value any)
	// 获取连接属性
	GetProperty(key IConnProperty) (value any)
	// 删除连接属性
	RemoveProperty(key IConnProperty)

	// 限流控制
	FlowControl() bool

	// 序列化
	ProtocolToByte(str proto.Message) []byte
	// 反序列化
	ByteToProtocol(byte []byte, target proto.Message) error
}

// 连接挂载的属性
type IConnProperty string

func (i IConnProperty) String() string {
	return string(i)
}
