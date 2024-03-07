package iface

import (
	"google.golang.org/protobuf/proto"
)

// 广播组管理器
type IBroadcastManager interface {
	// 新建广播组
	NewBroadcastGroup() IBroadcastGroup
	// 获取全局广播组
	GetGlobalBroadcastGroup() IBroadcastGroup
	// 根据组Id获取广播组
	GetBroadcastGroupById(groupId int64) (IBroadcastGroup, bool)
	// 根据组Id删除广播组(解散广播组)
	DelBroadcastGroupById(groupId int64)
	// 根据连接Id获取广播组
	GetBroadcastGroupByConnId(connId int) ([]IBroadcastGroup, bool)
	// 根据连接Id设置广播组(加入广播组)
	SetBroadcastGroupByConnId(connId int, broadcastGroup IBroadcastGroup)
	// 根据连接Id删除广播组(退出广播组)
	DelBroadcastGroupByConnId(connId int, broadcastGroup IBroadcastGroup)
}

// 广播组
type IBroadcastGroup interface {
	// 获取组Id
	GetGroupId() int64
	// 设置广播对象
	SetBroadcastTarget(connId int)
	// 获取广播对象
	GetBroadcastTarget(connId int) bool
	// 删除一个广播对象
	DelBroadcastTarget(connId int)
	// 清空组内所有广播对象
	ClearAllBroadcastTarget()
	// 广播所有对象
	BroadcastAllTargets(msgId int32, data proto.Message)
	// TODO 广播历史记录
}
