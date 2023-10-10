package iface

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

// 通知管理器
type INotifyManager interface {
	// 设置通知组
	SetNotify(notify INotify)
	// 获取通知组
	GetNotify(notifyID uint32) (any, bool)
	// 根据组ID删除指定通知
	DelNotifyByID(notifyID uint32)
	// 向通知组内所有对象广播信息
	SendNotifyData(notifyID uint32, msgID pb.MSgID, data proto.Message)
}

// 通知组
type INotify interface {
	// 获取通知对象
	GetNotifyID() uint32
	// 设置通知对象
	SetNotifyTarget(conn IConnection)
	// 获取通知对象列表
	GetNotifyTarget(connID uint32) (IConnection, bool)
	// 删除一个通知对象
	DelNotifyTarget(connID uint32)

	NotifyAllTargets(msgID pb.MSgID, data proto.Message)
}
