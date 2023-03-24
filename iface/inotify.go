package iface

import pb "github.com/451008604/socketServerFrame/proto/bin"

// 通知管理器
type INotifyManager interface {
	// 获取通知管理器
	GetNotifyGroup() map[string]INotify
	// 根据ID删除指定通知
	DelNotifyGroupByID(notifyID string)
	// 向通知列表内所有对象广播信息
	SendNotifyData(notifyID string, msgID pb.MessageID, data []byte)
}

// 通知
type INotify interface {
	// // 设置通知ID
	// SetNotifyID()
	// // 获取通知ID
	// GetNotifyID()
	// 设置通知对象
	SetNotifyTarget(conn IConnection)
	// 获取通知对象列表
	GetNotifyTargets() (connList []IConnection)
	// 删除一个通知对象
	DelNotifyTarget(conn IConnection)
	// 查询通知对象是否已存在
	HasNotifyTarget(conn IConnection) bool
}
