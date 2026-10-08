package nets

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"sync"
)

var messagePool = sync.Pool{New: func() any { return &Message{} }}

func GetMessage() *Message { return messagePool.Get().(*Message) }
func PutMessage(msg IMessage) {
	if m, ok := msg.(*Message); ok && m != nil {
		m.Reset()
		messagePool.Put(m)
	}
}

// Message keeps the uint16 public API while exposing a valid protobuf envelope:
// uint32 msg_id = 1; bytes data = 2. DataLen is framing metadata, not a wire field.
type Message struct {
	Id      uint16 `json:"msg_id"`
	Data    []byte `json:"data"`
	DataLen uint16 `json:"-"`
	unknown []byte
	wireID  uint32
}

func (m *Message) Reset()      { *m = Message{} }
func (*Message) ProtoMessage() {}
func (m *Message) GetDataLen() uint16 {
	if m.DataLen != 0 {
		return m.DataLen
	}
	return uint16(len(m.Data))
}
func (m *Message) GetMsgId() uint16    { return m.Id }
func (m *Message) GetData() []byte     { return m.Data }
func (m *Message) SetData(data []byte) { m.Data = data }

var messageDescriptor = buildMessageDescriptor()

func buildMessageDescriptor() protoreflect.MessageDescriptor {
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax: proto.String("proto3"), Name: proto.String("nets/envelope.proto"), Package: proto.String("nets"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Message"),
			Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("msg_id"), JsonName: proto.String("msg_id"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum()},
				{Name: proto.String("data"), Number: proto.Int32(2), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum()},
			},
		}},
	}, nil)
	if err != nil {
		panic(err)
	}
	return file.Messages().Get(0)
}

// The reflection adapter reads and writes the public fields directly. An embedded
// nil proto.Message cannot provide reflection, and a detached snapshot loses decoding.
type messageReflection struct {
	protoreflect.Message
	target *Message
}
type messageType struct{ protoreflect.MessageType }

func (t messageType) New() protoreflect.Message  { return (&Message{}).ProtoReflect() }
func (t messageType) Zero() protoreflect.Message { return (*Message)(nil).ProtoReflect() }
func (m *Message) ProtoReflect() protoreflect.Message {
	return messageReflection{Message: dynamicpb.NewMessage(messageDescriptor), target: m}
}
func (m messageReflection) Type() protoreflect.MessageType       { return messageType{m.Message.Type()} }
func (m messageReflection) New() protoreflect.Message            { return (&Message{}).ProtoReflect() }
func (m messageReflection) Interface() protoreflect.ProtoMessage { return m.target }
func (m messageReflection) IsValid() bool                        { return m.target != nil }
func (m messageReflection) id() uint32 {
	if m.target == nil {
		return 0
	}
	if m.target.wireID > 65535 && uint16(m.target.wireID) == m.target.Id {
		return m.target.wireID
	}
	return uint32(m.target.Id)
}
func (m messageReflection) Get(field protoreflect.FieldDescriptor) protoreflect.Value {
	_ = m.Message.Get(field) // Validate the descriptor.
	if m.target == nil {
		return field.Default()
	}
	if field.Number() == 1 {
		return protoreflect.ValueOfUint32(m.id())
	}
	return protoreflect.ValueOfBytes(m.target.Data)
}
func (m messageReflection) Has(field protoreflect.FieldDescriptor) bool {
	_ = m.Message.Has(field)
	if m.target == nil {
		return false
	}
	if field.Number() == 1 {
		return m.id() != 0
	}
	return len(m.target.Data) != 0
}
func (m messageReflection) Set(field protoreflect.FieldDescriptor, value protoreflect.Value) {
	if m.target == nil {
		panic("mutation of nil Message")
	}
	m.Message.Set(field, value) // Validate scalar type and field ownership.
	if field.Number() == 1 {
		id := uint32(value.Uint())
		m.target.Id, m.target.wireID = uint16(id), id
	} else {
		m.target.Data = value.Bytes()
	}
}
func (m messageReflection) Clear(field protoreflect.FieldDescriptor) {
	if m.target == nil {
		panic("mutation of nil Message")
	}
	m.Message.Clear(field)
	if field.Number() == 1 {
		m.target.Id, m.target.wireID = 0, 0
	} else {
		m.target.Data = nil
	}
}
func (m messageReflection) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	fields := messageDescriptor.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if m.Has(field) && !f(field, m.Get(field)) {
			return
		}
	}
}
func (m messageReflection) GetUnknown() protoreflect.RawFields {
	if m.target == nil {
		return nil
	}
	return m.target.unknown
}
func (m messageReflection) SetUnknown(fields protoreflect.RawFields) {
	if m.target == nil {
		panic("mutation of nil Message")
	}
	m.target.unknown = fields
}
