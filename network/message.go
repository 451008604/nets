package network

import "github.com/451008604/nets/iface"

type message struct {
	id      uint16 // 消息ID
	dataLen uint16 // 消息长度
	data    []byte // 消息内容
}

func NewMsgPackage(id int32, data []byte) iface.IMessage {
	return &message{
		id:      uint16(id),
		dataLen: uint16(len(data)),
		data:    data,
	}
}

func (m *message) GetDataLen() uint16 {
	return m.dataLen
}

func (m *message) SetDataLen(u uint16) {
	m.dataLen = u
}

func (m *message) GetMsgId() uint16 {
	return m.id
}

func (m *message) SetMsgId(u uint16) {
	m.id = u
}

func (m *message) GetData() []byte {
	return m.data
}

func (m *message) SetData(bytes []byte) {
	m.data = bytes
}
