package nets

import (
	"google.golang.org/protobuf/proto"
	"sync"
)

var messagePool = sync.Pool{
	New: func() any { return &Message{} },
}

func GetMessage() *Message {
	return messagePool.Get().(*Message)
}

func PutMessage(m IMessage) {
	if msg, ok := m.(*Message); ok {
		msg.Message = nil
		msg.Id = 0
		msg.DataLen = 0
		msg.Data = nil
		messagePool.Put(msg)
	}
}

type Message struct {
	proto.Message `json:"-"`
	Id            uint16 `protobuf:"bytes,1,opt,name=msg_id,proto3" json:"msg_id"` // Message ID / 消息Id
	Data          []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data"`     // Message Content / 消息内容
	DataLen       uint16 `json:"-"`                                                // Message Length / 消息长度
}

func (m *Message) GetDataLen() uint16 {
	if m.DataLen != 0 {
		return m.DataLen
	}
	return uint16(len(m.Data))
}

func (m *Message) GetMsgId() uint16 {
	return m.Id
}

func (m *Message) GetData() []byte {
	return m.Data
}

func (m *Message) SetData(bytes []byte) {
	m.Data = bytes
}
