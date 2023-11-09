package iface

import (
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
)

type IConnection interface {
	// 启动连接
	Start(readerHandler func(), writerHandler func(data []byte))
	// 停止连接
	Stop()

	// 启动接收消息协程
	StartReader()
	// 启动发送消息协程
	StartWriter(data []byte)

	// 获取当前连接ID
	GetConnID() int
	// 获取客户端地址信息
	RemoteAddrStr() string

	// 发送消息给客户端
	SendMsg(msgId pb.MSgID, msgData proto.Message)

	// 设置连接属性
	SetProperty(key string, value interface{})
	// 获取连接属性
	GetProperty(key string) (value interface{})
	// 删除连接属性
	RemoveProperty(key string)

	// 绑定 conn 对应的 player 对象
	SetPlayer(player interface{})
	// Deprecated: 不建议直接调用。应通过`logic.GetPlayer`获取 conn 绑定的 player 对象
	GetPlayer() interface{}

	SetNotifyGroupCh(notifyGroupCh IBroadcastData)
	// 加入广播组
	JoinBroadcastGroup(conn IConnection, group IBroadcast)
	// 根据组ID退出广播组
	ExitBroadcastGroupByID(groupID int64)
	// 退出所有广播组
	ExitAllBroadcastGroup()

	// 协议转字节
	ProtocolToByte(str proto.Message) []byte
	// 字节转协议
	ByteToProtocol(byte []byte, target proto.Message) error
}
