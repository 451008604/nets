package network

import pb "github.com/451008604/nets/proto/bin"

type Message struct {
	totalLen uint16 // 消息总长度
	id       uint16 // 消息ID
	dataLen  uint16 // 消息长度
	data     []byte // 消息内容
}

// 新建消息包
func NewMsgPackage(id pb.MSgID, data []byte) *Message {
	return &Message{
		id:       uint16(id),
		dataLen:  uint16(len(data)),
		data:     data,
		totalLen: uint16(len(data) + 6),
	}
}

func (m *Message) GetDataLen() uint16 {
	return m.dataLen
}

func (m *Message) SetDataLen(u uint16) {
	m.dataLen = u
}

func (m *Message) GetMsgId() uint16 {
	return m.id
}

func (m *Message) SetMsgId(u uint16) {
	m.id = u
}

func (m *Message) GetData() []byte {
	return m.data
}

func (m *Message) SetData(bytes []byte) {
	m.data = bytes
}

func (m *Message) GetTotalLen() uint16 {
	return m.totalLen
}

func (m *Message) SetTotalLen(totalLen uint16) {
	m.totalLen = totalLen
}
