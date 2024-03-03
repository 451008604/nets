package iface

import (
	"google.golang.org/protobuf/proto"
)

type IConnection interface {
	// 启动连接(通过connmanager调用)
	Start(readerHandler func(), writerHandler func(data []byte))
	// 停止连接(通过connmanager调用)
	Stop()

	// 启动接收消息协程
	StartReader()
	// 启动发送消息协程
	StartWriter(data []byte)

	// 获取当前连接Id
	GetConnId() int
	// 获取客户端地址信息
	RemoteAddrStr() string

	// 发送消息给客户端
	SendMsg(msgId int32, msgData proto.Message)

	// 设置连接属性
	SetProperty(key string, value any)
	// 获取连接属性
	GetProperty(key string) (value any)
	// 删除连接属性
	RemoveProperty(key string)

	// 加入广播组
	JoinBroadcastGroup(conn IConnection, groupId int64)
	// 根据组Id退出广播组
	ExitBroadcastGroup(groupId int64)
	// 退出所有广播组
	ExitAllBroadcastGroup()

	// 序列化
	ProtocolToByte(str proto.Message) []byte
	// 反序列化
	ByteToProtocol(byte []byte, target proto.Message) error
}
