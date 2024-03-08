package iface

import "google.golang.org/protobuf/proto"

// 广播组管理器
type IBroadcastManager interface {
	// 新建广播组
	NewBroadcastGroup() IMutexArray
	// 获取全局广播组
	GetGlobalBroadcastGroup() IMutexArray
	// 加入广播组
	JoinBroadcastGroup(groupId int, connId int)
	// 退出广播组
	ExitBroadcastGroup(groupId int, connId int)
	// 广播所有目标
	SendBroadcastAllTargets(groupId int, msgId int32, data proto.Message)
	// 根据组Id获取广播组
	GetBroadcastGroupByGroupId(groupId int) (IMutexArray, bool)
	// 根据连接Id获取广播组
	GetBroadcastGroupByConnId(connId int) (IMutexArray, bool)
}

// 广播组
type IMutexArray interface {
	// 追加一个元素
	Append(id int)
	// 删除一个元素
	Remove(id int)
	// 获取所有元素
	GetArray() []int
	// TODO 广播历史记录
}
