package iface

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

type IConnection interface {
	// 启动连接
	Start(writerHandler func(data []byte))
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
	SendMsg(msgId pb.MSG_ID, msgData proto.Message)

	// 设置连接属性
	SetProperty(key string, value interface{})
	// 获取连接属性
	GetProperty(key string) (value interface{})
	// 删除连接属性
	RemoveProperty(key string)

	SetPlayer(player interface{})
	// Deprecated: 不建议直接调用，应通过`logic.GetPlayer`获取 conn 绑定的 player 实例化对象
	GetPlayer() interface{}

	// 协议转字节
	ProtocolToByte(str proto.Message) []byte
	// 字节转协议
	ByteToProtocol(byte []byte, target proto.Message) error
}
