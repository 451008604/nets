package nets

import (
	"context"
	"google.golang.org/protobuf/proto"
	"net"
)

type IConnection interface {
	// Start Connection (called by connmanager) / 启动连接(通过connmanager调用)
	Open()
	// Stop Connection (called by connmanager) / 停止连接(通过connmanager调用)
	Close()

	// Get Real Connection / 获取真实连接
	GetNetConn() net.Conn
	// Connection Context / 连接上下文
	ConnCtx() context.Context

	// Start Message Receiving Goroutine / 启动接收消息协程
	StartReader() bool
	// Start Message Sending Goroutine / 启动发送消息协程
	StartWriter(data []byte) bool
	// Execute Task / 执行任务
	DoTask(task func()) bool

	// Get Current Connection ID / 获取当前连接Id
	GetConnId() string
	// Get Client Address Info / 获取客户端地址信息
	RemoteAddrStr() string
	// Get Whether Connection is Closed / 获取连接是否已关闭
	IsClose() bool
	// Get Connection Bound Property / 获取连接绑定的属性
	GetProperty(key string) any
	// Set Connection Bound Property / 设置连接绑定的属性
	SetProperty(key string, value any)
	// Remove Connection Bound Property / 移除连接绑定的属性
	RemoveProperty(key string)

	// Send Message to Client / 发送消息给客户端
	SendMsg(msgId int32, msgData proto.Message)

	// Rate Limiting Control / 限流控制
	FlowControl() bool

	// Serialize / 序列化
	ProtocolToByte(str proto.Message) []byte
	// Deserialize / 反序列化
	ByteToProtocol(byte []byte, target proto.Message) error
}
