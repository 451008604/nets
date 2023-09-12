package iface

type IMessage interface {
	// 获取消息总长度
	GetTotalLen() uint16
	// 设置消息内容
	SetTotalLen(totalLen uint16)

	// 获取消息ID
	GetMsgId() uint16
	// 设置消息ID
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
