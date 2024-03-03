package iface

// 定义消息模板
type IMessage interface {
	// 获取消息Id
	GetMsgId() uint16
	// 设置消息Id
	SetMsgId(id uint16)

	// 获取消息长度
	GetDataLen() uint16
	// 设置消息长度
	SetDataLen(uint16)

	// 获取消息内容
	GetData() []byte
	// 设置消息内容
	SetData([]byte)
}
