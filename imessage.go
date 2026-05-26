package nets

// Define Message Template / 定义消息模板
type IMessage interface {
	// Get Message ID / 获取消息Id
	GetMsgId() uint16
	// Get Message Length / 获取消息长度
	GetDataLen() uint16
	// Get Message Content / 获取消息内容
	GetData() []byte
	// Set Message Content / 设置消息内容
	SetData([]byte)
}
