package nets

import (
	"google.golang.org/protobuf/proto"
)

type Message struct {
	proto.Message `json:"-"`
	Id            uint16 `protobuf:"bytes,1,opt,name=msg_id,proto3" json:"msg_id"` // 消息Id
	Data          string `protobuf:"bytes,2,opt,name=data,proto3" json:"data"`     // 消息内容
	DataLen       uint16 `json:"-"`                                                // 消息长度
}

func NewMsgPackage(id int32, data []byte) IMessage {
	return &Message{
		Id:      uint16(id),
		DataLen: uint16(len(data)),
		Data:    string(data),
	}
}

func (m *Message) GetDataLen() uint16 {
	return m.DataLen
}

func (m *Message) GetMsgId() uint16 {
	return m.Id
}

func (m *Message) GetData() []byte {
	return []byte(m.Data)
}

func (m *Message) SetData(bytes []byte) {
	m.Data = string(bytes)
}
