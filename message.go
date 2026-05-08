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
	Id            uint16 `protobuf:"bytes,1,opt,name=msg_id,proto3" json:"msg_id"` // 消息Id
	Data          []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data"`     // 消息内容
	DataLen       uint16 `json:"-"`                                                // 消息长度
}

func NewMsgPackage(id int32, data []byte) IMessage {
	return &Message{
		Id:      uint16(id),
		DataLen: uint16(len(data)),
		Data:    data,
	}
}

func (m *Message) GetDataLen() uint16 {
	return m.DataLen
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
