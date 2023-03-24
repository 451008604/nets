// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.1
// source: MsgID.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type MessageID int32

const (
	MessageID_UNKNOWN MessageID = 0
	MessageID_PING    MessageID = 10000
	MessageID_Login   MessageID = 10001
)

// Enum value maps for MessageID.
var (
	MessageID_name = map[int32]string{
		0:     "UNKNOWN",
		10000: "PING",
		10001: "Login",
	}
	MessageID_value = map[string]int32{
		"UNKNOWN": 0,
		"PING":    10000,
		"Login":   10001,
	}
)

func (x MessageID) Enum() *MessageID {
	p := new(MessageID)
	*p = x
	return p
}

func (x MessageID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MessageID) Descriptor() protoreflect.EnumDescriptor {
	return file_MsgID_proto_enumTypes[0].Descriptor()
}

func (MessageID) Type() protoreflect.EnumType {
	return &file_MsgID_proto_enumTypes[0]
}

func (x MessageID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MessageID.Descriptor instead.
func (MessageID) EnumDescriptor() ([]byte, []int) {
	return file_MsgID_proto_rawDescGZIP(), []int{0}
}

var File_MsgID_proto protoreflect.FileDescriptor

var file_MsgID_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x4d, 0x73, 0x67, 0x49, 0x44, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70,
	0x62, 0x2a, 0x2f, 0x0a, 0x09, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x44, 0x12, 0x0b,
	0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x04, 0x50,
	0x49, 0x4e, 0x47, 0x10, 0x90, 0x4e, 0x12, 0x0a, 0x0a, 0x05, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x10,
	0x91, 0x4e, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_MsgID_proto_rawDescOnce sync.Once
	file_MsgID_proto_rawDescData = file_MsgID_proto_rawDesc
)

func file_MsgID_proto_rawDescGZIP() []byte {
	file_MsgID_proto_rawDescOnce.Do(func() {
		file_MsgID_proto_rawDescData = protoimpl.X.CompressGZIP(file_MsgID_proto_rawDescData)
	})
	return file_MsgID_proto_rawDescData
}

var file_MsgID_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_MsgID_proto_goTypes = []interface{}{
	(MessageID)(0), // 0: pb.MessageID
}
var file_MsgID_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_MsgID_proto_init() }
func file_MsgID_proto_init() {
	if File_MsgID_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_MsgID_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_MsgID_proto_goTypes,
		DependencyIndexes: file_MsgID_proto_depIdxs,
		EnumInfos:         file_MsgID_proto_enumTypes,
	}.Build()
	File_MsgID_proto = out.File
	file_MsgID_proto_rawDesc = nil
	file_MsgID_proto_goTypes = nil
	file_MsgID_proto_depIdxs = nil
}
