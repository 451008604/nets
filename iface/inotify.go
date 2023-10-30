package iface

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

// 广播组管理器
type IBroadcastManager interface {
	// 新建广播组
	NewBroadcastGroup() IBroadcast
	// 获取全局广播组
	GetGlobalBroadcast() IBroadcast
	// 获取广播组
	GetBroadcast(groupID int64) (any, bool)
	// 根据组ID删除指定广播组
	DelBroadcastByID(groupID int64)
	// 向组内所有对象广播信息
	SendBroadcastData(groupID int64, connID int, msgID pb.MSgID, data proto.Message)
}

// 广播组
type IBroadcast interface {
	// 获取组ID
	GetGroupID() int64
	// 设置广播对象
	SetBroadcastTarget(conn IConnection)
	// 获取广播对象
	GetBroadcastTarget(connID int) (IConnection, bool)
	// 删除一个广播对象
	DelBroadcastTarget(connID int)
	// 广播所有对象
	BroadcastAllTargets(connID int, msgID pb.MSgID, data proto.Message)
}

// 广播数据
type IBroadcastData interface {
	GroupID() int64
	SetGroupID(groupID int64)
	MsgID() pb.MSgID
	SetMsgID(msgID pb.MSgID)
	MsgData() proto.Message
	SetMsgData(msgData proto.Message)
}
