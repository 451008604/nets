// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v4.22.2
// source: id.proto

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

type MSgID int32

const (
	MSgID_None             MSgID = 0
	MSgID_PlayerLogin_Req  MSgID = 1001 // PlayerLoginReq  玩家登录请求
	MSgID_PlayerLogin_Res  MSgID = 1002 // PlayerLoginRes  玩家登录回应
	MSgID_Heartbeat_Req    MSgID = 1003 // HeartbeatReq    心跳请求
	MSgID_Heartbeat_Res    MSgID = 1004 // HeartbeatRes    心跳响应
	MSgID_Broadcast_Req    MSgID = 1005 // BroadcastRequest
	MSgID_Broadcast_Res    MSgID = 1006 // BroadcastResponse
	MSgID_ServerErr_Notify MSgID = 1007
)

// Enum value maps for MSgID.
var (
	MSgID_name = map[int32]string{
		0:    "None",
		1001: "PlayerLogin_Req",
		1002: "PlayerLogin_Res",
		1003: "Heartbeat_Req",
		1004: "Heartbeat_Res",
		1005: "Broadcast_Req",
		1006: "Broadcast_Res",
		1007: "ServerErr_Notify",
	}
	MSgID_value = map[string]int32{
		"None":             0,
		"PlayerLogin_Req":  1001,
		"PlayerLogin_Res":  1002,
		"Heartbeat_Req":    1003,
		"Heartbeat_Res":    1004,
		"Broadcast_Req":    1005,
		"Broadcast_Res":    1006,
		"ServerErr_Notify": 1007,
	}
)

func (x MSgID) Enum() *MSgID {
	p := new(MSgID)
	*p = x
	return p
}

func (x MSgID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MSgID) Descriptor() protoreflect.EnumDescriptor {
	return file_id_proto_enumTypes[0].Descriptor()
}

func (MSgID) Type() protoreflect.EnumType {
	return &file_id_proto_enumTypes[0]
}

func (x MSgID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MSgID.Descriptor instead.
func (MSgID) EnumDescriptor() ([]byte, []int) {
	return file_id_proto_rawDescGZIP(), []int{0}
}

var File_id_proto protoreflect.FileDescriptor

var file_id_proto_rawDesc = []byte{
	0x0a, 0x08, 0x69, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x2a, 0xa4,
	0x01, 0x0a, 0x05, 0x4d, 0x53, 0x67, 0x49, 0x44, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x6f, 0x6e, 0x65,
	0x10, 0x00, 0x12, 0x14, 0x0a, 0x0f, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69,
	0x6e, 0x5f, 0x52, 0x65, 0x71, 0x10, 0xe9, 0x07, 0x12, 0x14, 0x0a, 0x0f, 0x50, 0x6c, 0x61, 0x79,
	0x65, 0x72, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x52, 0x65, 0x73, 0x10, 0xea, 0x07, 0x12, 0x12,
	0x0a, 0x0d, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x5f, 0x52, 0x65, 0x71, 0x10,
	0xeb, 0x07, 0x12, 0x12, 0x0a, 0x0d, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x5f,
	0x52, 0x65, 0x73, 0x10, 0xec, 0x07, 0x12, 0x12, 0x0a, 0x0d, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63,
	0x61, 0x73, 0x74, 0x5f, 0x52, 0x65, 0x71, 0x10, 0xed, 0x07, 0x12, 0x12, 0x0a, 0x0d, 0x42, 0x72,
	0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x5f, 0x52, 0x65, 0x73, 0x10, 0xee, 0x07, 0x12, 0x15,
	0x0a, 0x10, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x45, 0x72, 0x72, 0x5f, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x79, 0x10, 0xef, 0x07, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x3b, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_id_proto_rawDescOnce sync.Once
	file_id_proto_rawDescData = file_id_proto_rawDesc
)

func file_id_proto_rawDescGZIP() []byte {
	file_id_proto_rawDescOnce.Do(func() {
		file_id_proto_rawDescData = protoimpl.X.CompressGZIP(file_id_proto_rawDescData)
	})
	return file_id_proto_rawDescData
}

var file_id_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_id_proto_goTypes = []interface{}{
	(MSgID)(0), // 0: pb.MSgID
}
var file_id_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_id_proto_init() }
func file_id_proto_init() {
	if File_id_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_id_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_id_proto_goTypes,
		DependencyIndexes: file_id_proto_depIdxs,
		EnumInfos:         file_id_proto_enumTypes,
	}.Build()
	File_id_proto = out.File
	file_id_proto_rawDesc = nil
	file_id_proto_goTypes = nil
	file_id_proto_depIdxs = nil
}
