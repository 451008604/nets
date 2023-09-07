package iface

import pb "github.com/451008604/socketServerFrame/proto/bin"

type IMessage interface {
	// 获取消息长度
	GetDataLen() uint32
	// 设置消息长度
	SetDataLen(uint32)

	// 获取消息ID
	GetMsgId() pb.MsgID
	// 设置消息ID
	SetMsgId(pb.MsgID)

	// 获取消息内容
	GetData() []byte
	// 设置消息内容
	SetData([]byte)
}
